'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';
import { createGalaxyShaderMaterial, GALAXY_SHADER_PRESETS } from '@/shaders/galaxyShaders';

interface StarCluster3DProps {
  ipSeed: number;
}

export function StarCluster3D({ ipSeed }: StarCluster3DProps) {
  const groupRef = useRef<THREE.Group>(null);
  
  const clusterParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    
    return {
      numStars: 800 + Math.floor(rng.random() * 400), // 800-1200 stars for richer appearance
      clusterRadius: 8 + rng.random() * 4, // 8-12 radius
      rotationSpeed: 0.01 + rng.random() * 0.02, // 0.01-0.03 cluster rotation
      isClockwise: rng.random() > 0.5,
      // Random cluster tilt for variety
      tilt: {
        x: (rng.random() - 0.5) * Math.PI * 0.5, // ±45 degrees
        y: (rng.random() - 0.5) * Math.PI * 0.3, // ±27 degrees
        z: (rng.random() - 0.5) * Math.PI * 0.4, // ±36 degrees
      },
      colors: {
        // Stellar evolution stages
        mainSequenceHot: new THREE.Color().setHSL(0.65 + rng.random() * 0.05, 0.9, 0.9), // Blue main sequence
        mainSequenceWarm: new THREE.Color().setHSL(0.15 + rng.random() * 0.05, 0.8, 0.8), // Yellow main sequence
        mainSequenceCool: new THREE.Color().setHSL(0.05 + rng.random() * 0.03, 0.7, 0.6), // Orange main sequence
        redGiant: new THREE.Color().setHSL(0.02 + rng.random() * 0.02, 0.8, 0.5), // Red giants
        whiteDrawf: new THREE.Color().setHSL(0.6 + rng.random() * 0.1, 0.3, 0.9), // White dwarfs
      }
    };
  }, [ipSeed]);

  const starData = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    const numStars = clusterParams.numStars;
    const positions = new Float32Array(numStars * 3);
    const colors = new Float32Array(numStars * 3);
    const sizes = new Float32Array(numStars);

    // GPU rotation attributes
    const originalPositions = new Float32Array(numStars * 3);
    const rotationCenters = new Float32Array(numStars * 3);
    const rotationAxes = new Float32Array(numStars * 3);
    const rotationSpeeds = new Float32Array(numStars);

    const clusterCenter = new THREE.Vector3(0, 0, 0);
    const rotationAxis = new THREE.Vector3(0, 1, 0);
    
    for (let i = 0; i < numStars; i++) {
      // King model distribution - realistic globular cluster profile
      const coreRadius = clusterParams.clusterRadius * 0.3; // Core radius
      const u1 = rng.random();
      const u2 = rng.random();
      
      // King model: higher density toward center, power-law falloff
      const radius = coreRadius * Math.sqrt(-Math.log(u1)) * (0.5 + u2 * 2.0);
      const actualRadius = Math.min(radius, clusterParams.clusterRadius);
      
      const theta = rng.random() * Math.PI * 2;
      const phi = Math.acos(2 * rng.random() - 1);
      
      const x = actualRadius * Math.sin(phi) * Math.cos(theta);
      const y = actualRadius * Math.sin(phi) * Math.sin(theta);
      const z = actualRadius * Math.cos(phi);
      
      positions[i * 3] = x;
      positions[i * 3 + 1] = y;
      positions[i * 3 + 2] = z;

      // Store original positions for GPU rotation
      originalPositions[i * 3] = x;
      originalPositions[i * 3 + 1] = y;
      originalPositions[i * 3 + 2] = z;

      // Set rotation center (cluster center)
      rotationCenters[i * 3] = clusterCenter.x;
      rotationCenters[i * 3 + 1] = clusterCenter.y;
      rotationCenters[i * 3 + 2] = clusterCenter.z;

      // Set rotation axis
      rotationAxes[i * 3] = rotationAxis.x;
      rotationAxes[i * 3 + 1] = rotationAxis.y;
      rotationAxes[i * 3 + 2] = rotationAxis.z;

      // Star cluster dynamics: slight differential rotation
      const normalizedRadius = actualRadius / clusterParams.clusterRadius;
      const baseSpeed = clusterParams.isClockwise ? clusterParams.rotationSpeed : -clusterParams.rotationSpeed;
      // Stars further out rotate slightly slower (not as pronounced as galaxies)
      const differentialSpeed = baseSpeed * (1.0 - normalizedRadius * 0.3);
      rotationSpeeds[i] = differentialSpeed;
      
      // Stellar evolution and classification
      const distanceFromCenter = actualRadius / clusterParams.clusterRadius;
      const stellarAge = rng.random(); // Represents stellar evolution stage
      const stellarMass = 0.3 + rng.random() * 2.0; // 0.3-2.3 solar masses
      
      let color;
      let stellarSize;
      
      // Stellar evolution based on mass and age
      if (stellarAge > 0.95 && stellarMass > 1.5) {
        // 5% - Evolved red giants (more common in older clusters)
        color = clusterParams.colors.redGiant;
        stellarSize = 2.0 + rng.random() * 1.5; // Large red giants
      } else if (stellarAge > 0.98 && stellarMass < 1.0) {
        // 2% - White dwarfs (end stage for low-mass stars)
        color = clusterParams.colors.whiteDrawf;
        stellarSize = 0.8 + rng.random() * 0.4; // Small but bright
      } else if (stellarMass > 1.8) {
        // Hot, massive main sequence stars (concentrated in core)
        const coreBonus = distanceFromCenter < 0.2 ? 1.5 : 1.0;
        color = clusterParams.colors.mainSequenceHot;
        stellarSize = (1.2 + stellarMass * 0.3) * coreBonus;
      } else if (stellarMass > 1.0) {
        // Warm main sequence stars
        color = clusterParams.colors.mainSequenceWarm;
        stellarSize = 0.8 + stellarMass * 0.4;
      } else {
        // Cool, low-mass main sequence stars (most common)
        color = clusterParams.colors.mainSequenceCool;
        stellarSize = 0.4 + stellarMass * 0.6;
      }
      
      // Add some brightness variation based on distance and stellar type
      const brightness = 0.7 + rng.random() * 0.6;
      const finalColor = color.clone().multiplyScalar(brightness);
      
      colors[i * 3] = finalColor.r;
      colors[i * 3 + 1] = finalColor.g;
      colors[i * 3 + 2] = finalColor.b;
      
      sizes[i] = stellarSize * (0.8 + rng.random() * 0.4);
    }
    
    return {
      positions,
      colors,
      sizes,
      originalPositions,
      rotationCenters,
      rotationAxes,
      rotationSpeeds
    };
  }, [ipSeed, clusterParams]);

  // Create Points with GPU shaders
  const clusterPoints = useMemo(() => {
    if (!starData.positions.length) return null;

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(starData.positions, 3));
    geometry.setAttribute('color', new THREE.Float32BufferAttribute(starData.colors, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(starData.sizes, 1));

    // Add rotation attributes for GPU-based rotation
    geometry.setAttribute('originalPosition', new THREE.Float32BufferAttribute(starData.originalPositions, 3));
    geometry.setAttribute('clusterCenter', new THREE.Float32BufferAttribute(starData.rotationCenters, 3));
    geometry.setAttribute('rotationAxis', new THREE.Float32BufferAttribute(starData.rotationAxes, 3));
    geometry.setAttribute('rotationSpeed', new THREE.Float32BufferAttribute(starData.rotationSpeeds, 1));

    const material = createGalaxyShaderMaterial(GALAXY_SHADER_PRESETS.cluster);

    return new THREE.Points(geometry, material);
  }, [starData]);

  useFrame((state) => {
    const time = state.clock.elapsedTime;

    // Update shader time uniform for GPU-based rotation and pulsing effects
    if (clusterPoints && clusterPoints.material instanceof THREE.ShaderMaterial) {
      clusterPoints.material.uniforms.time.value = time;
    }

    // Apply cluster tilt and gentle precession (stellar cluster dynamics)
    if (groupRef.current) {
      // Apply base tilt from cluster parameters
      groupRef.current.rotation.x = clusterParams.tilt.x + Math.sin(time * 0.08) * 0.05;
      groupRef.current.rotation.y = clusterParams.tilt.y + Math.cos(time * 0.06) * 0.03;
      groupRef.current.rotation.z = clusterParams.tilt.z + Math.sin(time * 0.12) * 0.04;
      
      // Subtle cluster "breathing" from stellar winds and internal dynamics
      const breathing = 1.0 + Math.sin(time * 0.3) * 0.02;
      groupRef.current.scale.setScalar(breathing);
    }
  });

  return (
    <group ref={groupRef}>
      {/* Star cluster with GPU differential rotation and stellar evolution */}
      {clusterPoints && (
        <primitive object={clusterPoints} />
      )}
    </group>
  );
}