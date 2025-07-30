'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import { Points, PointMaterial } from '@react-three/drei';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

interface Supercluster3DProps {
  ipSeed: number;
}

export function Supercluster3D({ ipSeed }: Supercluster3DProps) {
  const groupRef = useRef<THREE.Group>(null);
  
  // Generate supercluster parameters
  const clusterParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    
    return {
      numGalaxyClusters: 3 + Math.floor(rng.random() * 5), // 3-7 galaxy clusters
      shellRadius: 20 + rng.random() * 15, // 20-35 radius
      colors: {
        dense: new THREE.Color().setHSL(0.05 + rng.random() * 0.1, 0.8, 0.8), // Orange-red
        medium: new THREE.Color().setHSL(0.15 + rng.random() * 0.1, 0.7, 0.7), // Yellow
        sparse: new THREE.Color().setHSL(0.6 + rng.random() * 0.1, 0.6, 0.5), // Blue
      }
    };
  }, [ipSeed]);

  // Generate galaxy cluster geometry
  const clusterGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    const totalPoints = 2000;
    const positions = new Float32Array(totalPoints * 3);
    const colors = new Float32Array(totalPoints * 3);
    
    let pointIndex = 0;
    
    for (let cluster = 0; cluster < clusterParams.numGalaxyClusters; cluster++) {
      const pointsInCluster = Math.floor(totalPoints / clusterParams.numGalaxyClusters);
      
      // Cluster center position
      const clusterRadius = clusterParams.shellRadius * (0.3 + rng.random() * 0.7);
      const clusterTheta = rng.random() * Math.PI * 2;
      const clusterPhi = Math.acos(2 * rng.random() - 1);
      
      const centerX = clusterRadius * Math.sin(clusterPhi) * Math.cos(clusterTheta);
      const centerY = clusterRadius * Math.sin(clusterPhi) * Math.sin(clusterTheta);
      const centerZ = clusterRadius * Math.cos(clusterPhi);
      
      for (let p = 0; p < pointsInCluster && pointIndex < totalPoints; p++) {
        // Galaxy position within cluster (roughly spherical with higher density at center)
        const galaxyRadius = Math.pow(rng.random(), 2) * 8;
        const galaxyTheta = rng.random() * Math.PI * 2;
        const galaxyPhi = Math.acos(2 * rng.random() - 1);
        
        const x = centerX + galaxyRadius * Math.sin(galaxyPhi) * Math.cos(galaxyTheta);
        const y = centerY + galaxyRadius * Math.sin(galaxyPhi) * Math.sin(galaxyTheta);
        const z = centerZ + galaxyRadius * Math.cos(galaxyPhi);
        
        positions[pointIndex * 3] = x;
        positions[pointIndex * 3 + 1] = y;
        positions[pointIndex * 3 + 2] = z;
        
        // Color based on density
        let color;
        if (galaxyRadius < 2) {
          color = clusterParams.colors.dense; // Dense core
        } else if (galaxyRadius < 5) {
          color = clusterParams.colors.medium; // Medium density
        } else {
          color = clusterParams.colors.sparse; // Sparse outskirts
        }
        
        colors[pointIndex * 3] = color.r;
        colors[pointIndex * 3 + 1] = color.g;
        colors[pointIndex * 3 + 2] = color.b;
        
        pointIndex++;
      }
    }
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    
    return geometry;
  }, [ipSeed, clusterParams]);

  // Animation loop
  useFrame((state) => {
    if (groupRef.current) {
      // Slow rotation around multiple axes
      groupRef.current.rotation.y = state.clock.elapsedTime * 0.02;
      groupRef.current.rotation.x = Math.sin(state.clock.elapsedTime * 0.01) * 0.1;
      groupRef.current.rotation.z = Math.cos(state.clock.elapsedTime * 0.015) * 0.05;
    }
  });

  return (
    <group ref={groupRef}>
      <Points geometry={clusterGeometry}>
        <PointMaterial
          vertexColors
          size={1.0}
          sizeAttenuation={true}
          transparent
          opacity={0.8}
        />
      </Points>
    </group>
  );
}