'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';
import { createGalaxyShaderMaterial, GALAXY_SHADER_PRESETS } from '@/shaders/galaxyShaders';

interface Supercluster3DProps {
  ipSeed: number;
}

export function Supercluster3D({ ipSeed }: Supercluster3DProps) {
  const groupRef = useRef<THREE.Group>(null);

  // Generate supercluster parameters
  const clusterParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);

    return {
      numGalaxyClusters: 50 + Math.floor(rng.random() * 20), // 50-70 galaxy clusters
      shellRadius: 20 + rng.random() * 25, // 20-45 radius
      colors: {
        dense: new THREE.Color().setHSL(0.05 + rng.random() * 0.1, 0.8, 0.8), // Orange-red
        medium: new THREE.Color().setHSL(0.15 + rng.random() * 0.1, 0.7, 0.7), // Yellow
        sparse: new THREE.Color().setHSL(0.6 + rng.random() * 0.1, 0.6, 0.5), // Blue
      }
    };
  }, [ipSeed]);

  // Generate galaxy cluster data with distance-based properties
  const galaxyClusterData = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    const totalPoints = 30000; // More points for richer appearance
    const positions: number[] = [];
    const colors: number[] = [];
    const sizes: number[] = [];

    // GPU rotation attributes
    const originalPositions: number[] = [];
    const clusterCenters: number[] = [];
    const rotationAxes: number[] = [];
    const rotationSpeeds: number[] = [];
    
    // Store cluster data for geometry setup
    const clusterCenterVectors: THREE.Vector3[] = [];
    const clusterRotationAxes: THREE.Vector3[] = [];
    const clusterRotationSpeeds: number[] = [];
    const galaxyClusterIds: number[] = [];

    for (let cluster = 0; cluster < clusterParams.numGalaxyClusters; cluster++) {
      const pointsInCluster = Math.floor(totalPoints / clusterParams.numGalaxyClusters);

      // Cluster center position
      const clusterRadius = clusterParams.shellRadius * (0.3 + rng.random() * 0.7);
      const clusterTheta = rng.random() * Math.PI * 2;
      const clusterPhi = Math.acos(2 * rng.random() - 1);

      const centerX = clusterRadius * Math.sin(clusterPhi) * Math.cos(clusterTheta);
      const centerY = clusterRadius * Math.sin(clusterPhi) * Math.sin(clusterTheta);
      const centerZ = clusterRadius * Math.cos(clusterPhi);

      const clusterCenter = new THREE.Vector3(centerX, centerY, centerZ);
      clusterCenterVectors.push(clusterCenter);

      // Generate rotation data for this cluster
      const rotationAxis = new THREE.Vector3(
        rng.random() - 0.5,
        rng.random() - 0.5,
        rng.random() - 0.5
      ).normalize();
      const rotationSpeed = (rng.random() - 0.5) * 0.05; // More visible rotation for testing
      
      clusterRotationAxes.push(rotationAxis);
      clusterRotationSpeeds.push(rotationSpeed);

      for (let p = 0; p < pointsInCluster; p++) {
        // Galaxy position within cluster (roughly spherical with higher density at center)
        const galaxyRadius = Math.pow(rng.random(), 2.5) * 10; // More concentrated
        const galaxyTheta = rng.random() * Math.PI * 2;
        const galaxyPhi = Math.acos(2 * rng.random() - 1);

        const x = centerX + galaxyRadius * Math.sin(galaxyPhi) * Math.cos(galaxyTheta);
        const y = centerY + galaxyRadius * Math.sin(galaxyPhi) * Math.sin(galaxyTheta);
        const z = centerZ + galaxyRadius * Math.cos(galaxyPhi);

        const galaxyPosition = new THREE.Vector3(x, y, z);
        positions.push(x, y, z);
        originalPositions.push(x, y, z);
        clusterCenters.push(centerX, centerY, centerZ);
        rotationAxes.push(rotationAxis.x, rotationAxis.y, rotationAxis.z);
        rotationSpeeds.push(rotationSpeed);
        galaxyClusterIds.push(cluster);

        // Calculate brightness based on distance from cluster center
        const distanceFromCenter = galaxyRadius;
        const maxDistance = 10.0; // Maximum expected distance
        const normalizedDistance = Math.min(distanceFromCenter / maxDistance, 1.0);
        const brightness = 1.0 - normalizedDistance * 0.8; // 0.2 to 1.0 range

        // Color variation with distance-based brightness
        const colorRng = new SeededRandom(ipSeed + x * 1000 + y * 1000 + z * 1000);
        let baseColor: THREE.Color;

        if (galaxyRadius < 2) {
          baseColor = clusterParams.colors.dense;
        } else if (galaxyRadius < 6) {
          baseColor = clusterParams.colors.medium;
        } else {
          baseColor = clusterParams.colors.sparse;
        }

        // Add color variation
        const colorVariation = colorRng.random();
        let finalColor: THREE.Color;

        if (colorVariation < 0.6) {
          // 60% chance: stay close to base color
          const hsl = { h: 0, s: 0, l: 0 };
          baseColor.getHSL(hsl);

          const hueShift = (colorRng.random() - 0.5) * 0.15;
          const satShift = (colorRng.random() - 0.5) * 0.3;
          const lightShift = (colorRng.random() - 0.5) * 0.2;

          finalColor = new THREE.Color().setHSL(
            (hsl.h + hueShift + 1) % 1,
            Math.max(0, Math.min(1, hsl.s + satShift)),
            Math.max(0, Math.min(1, hsl.l + lightShift))
          );
        } else if (colorVariation < 0.85) {
          // 25% chance: warmer colors for variety
          finalColor = new THREE.Color().setHSL(
            0.1 + colorRng.random() * 0.2, // Orange to yellow
            0.7 + colorRng.random() * 0.3,
            0.5 + colorRng.random() * 0.3
          );
        } else {
          // 15% chance: cooler colors
          finalColor = new THREE.Color().setHSL(
            0.5 + colorRng.random() * 0.3, // Cyan to purple
            0.6 + colorRng.random() * 0.4,
            0.4 + colorRng.random() * 0.4
          );
        }

        // Apply brightness to color
        const brightColor = finalColor.multiplyScalar(brightness * 1.8);
        colors.push(brightColor.r, brightColor.g, brightColor.b);

        // Size based on brightness and distance
        const baseSize = 0.5 + rng.random() * 1.0;
        const pointSize = baseSize * (0.3 + brightness * 1.2); // Scale for points
        sizes.push(pointSize);
      }
    }

    return { 
      positions, 
      colors, 
      sizes, 
      originalPositions,
      clusterCenters,
      rotationAxes,
      rotationSpeeds,
      clusterCenterVectors, 
      clusterRotationAxes, 
      clusterRotationSpeeds, 
      galaxyClusterIds
    };
  }, [ipSeed, clusterParams]);

  // Create points mesh with custom glowing shader
  const galaxyPoints = useMemo(() => {
    if (galaxyClusterData.positions.length === 0) return null;

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(galaxyClusterData.positions, 3));
    geometry.setAttribute('color', new THREE.Float32BufferAttribute(galaxyClusterData.colors, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(galaxyClusterData.sizes, 1));
    
    // Add rotation attributes for GPU-based rotation
    geometry.setAttribute('originalPosition', new THREE.Float32BufferAttribute(galaxyClusterData.originalPositions, 3));
    geometry.setAttribute('clusterCenter', new THREE.Float32BufferAttribute(galaxyClusterData.clusterCenters, 3));
    geometry.setAttribute('rotationAxis', new THREE.Float32BufferAttribute(galaxyClusterData.rotationAxes, 3));
    geometry.setAttribute('rotationSpeed', new THREE.Float32BufferAttribute(galaxyClusterData.rotationSpeeds, 1));

    // Use shared galaxy shader with supercluster preset (rotation enabled)
    const material = createGalaxyShaderMaterial(GALAXY_SHADER_PRESETS.supercluster);

    return new THREE.Points(geometry, material);
  }, [galaxyClusterData]);

  // Animation loop - now only updates shader time uniform
  useFrame((state) => {
    // Update shader time uniform for GPU-based rotation and pulsing effects
    if (galaxyPoints && galaxyPoints.material instanceof THREE.ShaderMaterial) {
      galaxyPoints.material.uniforms.time.value = state.clock.elapsedTime;
    }
  });

  return (
    <group ref={groupRef}>
      {/* Glowing supercluster galaxies */}
      {galaxyPoints && (
        <primitive object={galaxyPoints} />
      )}
    </group>
  );
}