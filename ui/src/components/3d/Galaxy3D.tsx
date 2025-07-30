'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import { Points, PointMaterial } from '@react-three/drei';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

interface Galaxy3DProps {
  ipSeed: number;
}

export function Galaxy3D({ ipSeed }: Galaxy3DProps) {
  const groupRef = useRef<THREE.Group>(null);
  const coreRef = useRef<THREE.Points>(null);
  const armsRef = useRef<THREE.Points[]>([]);
  
  // Generate galaxy parameters based on IP seed
  const galaxyParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    
    return {
      numArms: 4 + Math.floor(rng.random() * 3), // 4-6 arms
      pointsPerArm: 1000 + Math.floor(rng.random() * 500), // 1000-1500 points per arm
      size: 0.8 + rng.random() * 0.4, // 0.8-1.2 size multiplier
      armTightness: 0.3 + rng.random() * 0.15, // 0.3-0.45 arm tightness
      spinSpeed: 0.02 + rng.random() * 0.01, // 0.02-0.03 rotation speed
      isClockwise: rng.random() > 0.5,
      colors: {
        core: new THREE.Color().setHSL(0.1 + rng.random() * 0.05, 0.8, 0.7), // Red-orange core
        stars: new THREE.Color().setHSL(0.15 + rng.random() * 0.05, 0.7, 0.8), // Yellow-white stars
        arms: new THREE.Color().setHSL(0.6 + rng.random() * 0.1, 0.8, 0.6), // Blue arms
        dust: new THREE.Color().setHSL(0.65 + rng.random() * 0.05, 0.5, 0.3), // Dark dust
      }
    };
  }, [ipSeed]);

  // Generate core geometry
  const coreGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    const corePoints = 200;
    const positions = new Float32Array(corePoints * 3);
    const colors = new Float32Array(corePoints * 3);
    
    for (let i = 0; i < corePoints; i++) {
      // Spherical distribution with higher density toward center
      const radius = Math.pow(rng.random(), 2) * 2 * galaxyParams.size;
      const theta = rng.random() * Math.PI * 2;
      const phi = Math.acos(2 * rng.random() - 1);
      
      positions[i * 3] = radius * Math.sin(phi) * Math.cos(theta);
      positions[i * 3 + 1] = radius * Math.sin(phi) * Math.sin(theta) * 0.3; // Flatten
      positions[i * 3 + 2] = radius * Math.cos(phi) * 0.3;
      
      // Core colors - brighter toward center
      const distanceFromCenter = radius / (2 * galaxyParams.size);
      const brightness = 1.0 - distanceFromCenter * 0.5;
      const color = galaxyParams.colors.core.clone().multiplyScalar(brightness);
      
      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;
    }
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    
    return geometry;
  }, [ipSeed, galaxyParams]);

  // Generate spiral arm geometries
  const armGeometries = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 100);
    const geometries = [];

    for (let arm = 0; arm < galaxyParams.numArms; arm++) {
      const positions = new Float32Array(galaxyParams.pointsPerArm * 3);
      const colors = new Float32Array(galaxyParams.pointsPerArm * 3);
      
      const armAngle = (arm * 2 * Math.PI) / galaxyParams.numArms;
      const armSeedBase = rng.random();
      const armSeedTwist = rng.random();
      
      for (let p = 0; p < galaxyParams.pointsPerArm; p++) {
        const progress = p / galaxyParams.pointsPerArm;
        
        // Spiral radius
        const r = progress * 8 * galaxyParams.size;
        
        // Spiral angle with wave variations
        const armWave = 0.2 * (0.8 + 0.4 * armSeedBase) * 
                       Math.sin(progress * 4) * (1.0 - progress * 0.5);
        const spiralTwist = Math.pow(progress, 0.6 + 0.2 * armSeedTwist);
        
        const twistDir = galaxyParams.isClockwise ? 1 : -1;
        const theta = armAngle + twistDir * r * galaxyParams.armTightness * spiralTwist + armWave;
        
        // Add some random scatter
        const scatter = (rng.random() - 0.5) * 0.5;
        const x = r * Math.cos(theta) + scatter;
        const z = r * Math.sin(theta) + scatter;
        const y = (rng.random() - 0.5) * 0.8 * (1.0 - progress); // Disk thickness
        
        positions[p * 3] = x;
        positions[p * 3 + 1] = y;
        positions[p * 3 + 2] = z;
        
        // Determine star type and color
        let color;
        const starType = rng.random();
        const distanceFromCenter = r / (8 * galaxyParams.size);
        
        if (p < galaxyParams.pointsPerArm / 6) {
          // Inner region - core stars
          color = galaxyParams.colors.core;
        } else if (starType > 0.8) {
          // Bright stars
          color = galaxyParams.colors.stars;
        } else if (starType > 0.5) {
          // Normal arm stars
          color = galaxyParams.colors.arms;
        } else {
          // Dust and dim stars
          const brightness = 0.5 - distanceFromCenter * 0.3;
          color = galaxyParams.colors.dust.clone().multiplyScalar(brightness);
        }
        
        colors[p * 3] = color.r;
        colors[p * 3 + 1] = color.g;
        colors[p * 3 + 2] = color.b;
      }
      
      const geometry = new THREE.BufferGeometry();
      geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
      geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
      
      geometries.push(geometry);
    }
    
    return geometries;
  }, [ipSeed, galaxyParams]);

  // Background particles for outer halo
  const haloGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 200);
    const particleCount = 300;
    const positions = new Float32Array(particleCount * 3);
    const colors = new Float32Array(particleCount * 3);
    
    for (let i = 0; i < particleCount; i++) {
      // Spherical halo distribution
      const radius = 10 + rng.random() * 15;
      const theta = rng.random() * Math.PI * 2;
      const phi = Math.acos(2 * rng.random() - 1);
      
      positions[i * 3] = radius * Math.sin(phi) * Math.cos(theta);
      positions[i * 3 + 1] = radius * Math.sin(phi) * Math.sin(theta) * 0.5;
      positions[i * 3 + 2] = radius * Math.cos(phi) * 0.5;
      
      const brightness = 0.2 + rng.random() * 0.3;
      const color = galaxyParams.colors.dust.clone().multiplyScalar(brightness);
      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;
    }
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    
    return geometry;
  }, [ipSeed, galaxyParams.colors.dust]);

  // Animation loop
  useFrame((state) => {
    if (groupRef.current) {
      // Rotate the entire galaxy
      const rotationSpeed = galaxyParams.isClockwise ? galaxyParams.spinSpeed : -galaxyParams.spinSpeed;
      groupRef.current.rotation.y += rotationSpeed;
      
      // Gentle precession
      groupRef.current.rotation.x = Math.sin(state.clock.elapsedTime * 0.1) * 0.1;
      groupRef.current.rotation.z = Math.cos(state.clock.elapsedTime * 0.15) * 0.05;
    }
    
    // Animate core with slight pulsing
    if (coreRef.current) {
      const scale = 1.0 + Math.sin(state.clock.elapsedTime * 2) * 0.05;
      coreRef.current.scale.setScalar(scale);
    }
  });

  return (
    <group ref={groupRef}>
      {/* Outer halo */}
      <Points geometry={haloGeometry}>
        <PointMaterial
          vertexColors
          size={0.5}
          sizeAttenuation={true}
          transparent
          opacity={0.4}
        />
      </Points>
      
      {/* Galaxy core */}
      <Points ref={coreRef} geometry={coreGeometry}>
        <PointMaterial
          vertexColors
          size={2.0}
          sizeAttenuation={true}
          transparent
          opacity={0.9}
        />
      </Points>
      
      {/* Spiral arms */}
      {armGeometries.map((geometry, index) => (
        <Points
          key={index}
          ref={(ref) => {
            if (ref) armsRef.current[index] = ref;
          }}
          geometry={geometry}
        >
          <PointMaterial
            vertexColors
            size={1.2}
            sizeAttenuation={true}
            transparent
            opacity={0.8}
          />
        </Points>
      ))}
    </group>
  );
}