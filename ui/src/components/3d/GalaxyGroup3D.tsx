'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import { Points, PointMaterial } from '@react-three/drei';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

interface GalaxyGroup3DProps {
  ipSeed: number;
}

export function GalaxyGroup3D({ ipSeed }: GalaxyGroup3DProps) {
  const groupRef = useRef<THREE.Group>(null);
  
  const groupParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    
    return {
      numGalaxies: 5 + Math.floor(rng.random() * 10), // 5-15 galaxies
      streamLength: 30,
      colors: {
        bright: new THREE.Color().setHSL(0.6 + rng.random() * 0.1, 0.9, 0.8),
        medium: new THREE.Color().setHSL(0.55 + rng.random() * 0.1, 0.7, 0.6),
        dim: new THREE.Color().setHSL(0.65 + rng.random() * 0.1, 0.5, 0.4),
      }
    };
  }, [ipSeed]);

  const streamGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    const totalPoints = 1500;
    const positions = new Float32Array(totalPoints * 3);
    const colors = new Float32Array(totalPoints * 3);
    
    for (let i = 0; i < totalPoints; i++) {
      // Create flowing stream pattern
      const t = i / totalPoints;
      const flow = Math.sin(t * Math.PI * 4 + rng.random() * Math.PI) * 3;
      
      const x = (t - 0.5) * groupParams.streamLength + flow;
      const y = Math.sin(t * Math.PI * 6) * 2 + (rng.random() - 0.5) * 2;
      const z = Math.cos(t * Math.PI * 6) * 2 + (rng.random() - 0.5) * 2;
      
      positions[i * 3] = x;
      positions[i * 3 + 1] = y;
      positions[i * 3 + 2] = z;
      
      // Color based on position in stream
      let color;
      if (t < 0.2 || t > 0.8) {
        color = groupParams.colors.dim;
      } else if (t < 0.4 || t > 0.6) {
        color = groupParams.colors.medium;
      } else {
        color = groupParams.colors.bright;
      }
      
      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;
    }
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    
    return geometry;
  }, [ipSeed, groupParams]);

  useFrame((state) => {
    if (groupRef.current) {
      groupRef.current.rotation.y = state.clock.elapsedTime * 0.03;
      groupRef.current.rotation.x = Math.sin(state.clock.elapsedTime * 0.02) * 0.2;
    }
  });

  return (
    <group ref={groupRef}>
      <Points geometry={streamGeometry}>
        <PointMaterial
          vertexColors
          size={0.8}
          sizeAttenuation={true}
          transparent
          opacity={0.7}
        />
      </Points>
    </group>
  );
}