'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import { Points, PointMaterial } from '@react-three/drei';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

interface StarCluster3DProps {
  ipSeed: number;
}

export function StarCluster3D({ ipSeed }: StarCluster3DProps) {
  const groupRef = useRef<THREE.Group>(null);
  
  const clusterParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    
    return {
      numStars: 300 + Math.floor(rng.random() * 200), // 300-500 stars
      clusterRadius: 8 + rng.random() * 4, // 8-12 radius
      colors: {
        hot: new THREE.Color().setHSL(0.6, 0.8, 0.9), // Blue-white
        warm: new THREE.Color().setHSL(0.15, 0.7, 0.8), // Yellow
        cool: new THREE.Color().setHSL(0.05, 0.6, 0.6), // Orange-red
      }
    };
  }, [ipSeed]);

  const starGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    const positions = new Float32Array(clusterParams.numStars * 3);
    const colors = new Float32Array(clusterParams.numStars * 3);
    
    for (let i = 0; i < clusterParams.numStars; i++) {
      // Spherical cluster with higher density toward center
      const radius = Math.pow(rng.random(), 1.5) * clusterParams.clusterRadius;
      const theta = rng.random() * Math.PI * 2;
      const phi = Math.acos(2 * rng.random() - 1);
      
      positions[i * 3] = radius * Math.sin(phi) * Math.cos(theta);
      positions[i * 3 + 1] = radius * Math.sin(phi) * Math.sin(theta);
      positions[i * 3 + 2] = radius * Math.cos(phi);
      
      // Stellar classification based on position (core stars are hotter)
      const distanceFromCenter = radius / clusterParams.clusterRadius;
      let color;
      
      if (distanceFromCenter < 0.3 && rng.random() > 0.7) {
        color = clusterParams.colors.hot; // Hot core stars
      } else if (rng.random() > 0.5) {
        color = clusterParams.colors.warm; // Main sequence stars
      } else {
        color = clusterParams.colors.cool; // Cool outer stars
      }
      
      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;
    }
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    
    return geometry;
  }, [ipSeed, clusterParams]);

  useFrame((state) => {
    if (groupRef.current) {
      // Gentle rotation and pulsing
      groupRef.current.rotation.y = state.clock.elapsedTime * 0.05;
      const scale = 1.0 + Math.sin(state.clock.elapsedTime * 0.5) * 0.05;
      groupRef.current.scale.setScalar(scale);
    }
  });

  return (
    <group ref={groupRef}>
      <Points geometry={starGeometry}>
        <PointMaterial
          vertexColors
          size={1.5}
          sizeAttenuation={true}
          transparent
          opacity={0.9}
        />
      </Points>
    </group>
  );
}