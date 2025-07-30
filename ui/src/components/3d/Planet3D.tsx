'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import { Sphere, Points, PointMaterial } from '@react-three/drei';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

interface Planet3DProps {
  ipSeed: number;
}

export function Planet3D({ ipSeed }: Planet3DProps) {
  const groupRef = useRef<THREE.Group>(null);
  const planetRef = useRef<THREE.Mesh>(null);
  const atmosphereRef = useRef<THREE.Mesh>(null);
  
  const planetParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    
    const typeRoll = rng.random();
    let type: 'rocky' | 'gas' | 'ice' | 'ocean';
    let baseColor: THREE.Color;
    let atmosphereColor: THREE.Color;
    
    if (typeRoll < 0.3) {
      type = 'rocky';
      baseColor = new THREE.Color().setHSL(0.08 + rng.random() * 0.05, 0.7, 0.5); // Brown-orange
      atmosphereColor = new THREE.Color().setHSL(0.15, 0.3, 0.8); // Thin atmosphere
    } else if (typeRoll < 0.5) {
      type = 'ocean';
      baseColor = new THREE.Color().setHSL(0.6 + rng.random() * 0.1, 0.8, 0.4); // Blue-green
      atmosphereColor = new THREE.Color().setHSL(0.6, 0.3, 0.9); // Thick atmosphere
    } else if (typeRoll < 0.7) {
      type = 'gas';
      baseColor = new THREE.Color().setHSL(0.15 + rng.random() * 0.1, 0.6, 0.7); // Yellow-orange
      atmosphereColor = new THREE.Color().setHSL(0.12, 0.4, 0.8); // Dense atmosphere
    } else {
      type = 'ice';
      baseColor = new THREE.Color().setHSL(0.55 + rng.random() * 0.1, 0.5, 0.8); // Light blue
      atmosphereColor = new THREE.Color().setHSL(0.65, 0.2, 0.9); // Thin atmosphere
    }
    
    return {
      type,
      size: 2.5 + rng.random() * 1.5, // 2.5-4.0 radius
      baseColor,
      atmosphereColor,
      rotationSpeed: 0.5 + rng.random() * 1.5, // 0.5-2.0 rotation speed
      hasRings: type === 'gas' && rng.random() > 0.6,
      cloudDensity: type === 'ocean' ? 0.8 : type === 'gas' ? 0.9 : 0.3,
    };
  }, [ipSeed]);

  // Generate cloud/atmosphere geometry
  const cloudGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 100);
    const cloudPoints = 200;
    const positions = new Float32Array(cloudPoints * 3);
    const colors = new Float32Array(cloudPoints * 3);
    
    for (let i = 0; i < cloudPoints; i++) {
      // Spherical distribution slightly above surface
      const radius = planetParams.size + 0.1 + rng.random() * 0.3;
      const theta = rng.random() * Math.PI * 2;
      const phi = Math.acos(2 * rng.random() - 1);
      
      positions[i * 3] = radius * Math.sin(phi) * Math.cos(theta);
      positions[i * 3 + 1] = radius * Math.sin(phi) * Math.sin(theta);
      positions[i * 3 + 2] = radius * Math.cos(phi);
      
      const opacity = 0.3 + rng.random() * 0.4;
      const color = planetParams.atmosphereColor.clone().multiplyScalar(opacity);
      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;
    }
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    
    return geometry;
  }, [ipSeed, planetParams]);

  // Generate surface features
  const surfaceGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 200);
    const featureCount = planetParams.type === 'gas' ? 50 : 100;
    const positions = new Float32Array(featureCount * 3);
    const colors = new Float32Array(featureCount * 3);
    
    for (let i = 0; i < featureCount; i++) {
      // Surface features at planet radius
      const theta = rng.random() * Math.PI * 2;
      const phi = Math.acos(2 * rng.random() - 1);
      
      positions[i * 3] = planetParams.size * Math.sin(phi) * Math.cos(theta);
      positions[i * 3 + 1] = planetParams.size * Math.sin(phi) * Math.sin(theta);
      positions[i * 3 + 2] = planetParams.size * Math.cos(phi);
      
      // Feature colors based on planet type
      let featureColor;
      if (planetParams.type === 'ocean') {
        featureColor = rng.random() > 0.7 ? 
          new THREE.Color(0.4, 0.8, 0.3) : // Land masses
          planetParams.baseColor.clone().multiplyScalar(0.8); // Ocean
      } else if (planetParams.type === 'rocky') {
        featureColor = rng.random() > 0.8 ?
          new THREE.Color(0.9, 0.9, 0.9) : // Ice caps
          planetParams.baseColor.clone().multiplyScalar(0.7 + rng.random() * 0.6);
      } else {
        featureColor = planetParams.baseColor.clone().multiplyScalar(0.8 + rng.random() * 0.4);
      }
      
      colors[i * 3] = featureColor.r;
      colors[i * 3 + 1] = featureColor.g;
      colors[i * 3 + 2] = featureColor.b;
    }
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    
    return geometry;
  }, [ipSeed, planetParams]);

  useFrame((state, delta) => {
    if (groupRef.current) {
      // Rotate the entire planet system
      groupRef.current.rotation.y += delta * planetParams.rotationSpeed * 0.1;
    }
    
    if (planetRef.current) {
      // Planet rotation
      planetRef.current.rotation.y += delta * planetParams.rotationSpeed;
    }
    
    if (atmosphereRef.current) {
      // Atmosphere rotates slightly faster
      atmosphereRef.current.rotation.y += delta * planetParams.rotationSpeed * 1.2;
    }
  });

  return (
    <group ref={groupRef}>
      {/* Planet core */}
      <Sphere 
        ref={planetRef}
        args={[planetParams.size, 32, 32]}
      >
        <meshLambertMaterial color={planetParams.baseColor} />
      </Sphere>
      
      {/* Surface features */}
      <Points geometry={surfaceGeometry}>
        <PointMaterial
          vertexColors
          size={0.3}
          sizeAttenuation={true}
          transparent
          opacity={0.8}
        />
      </Points>
      
      {/* Atmosphere/clouds */}
      <Sphere 
        ref={atmosphereRef}
        args={[planetParams.size + 0.2, 16, 16]}
      >
        <meshBasicMaterial 
          color={planetParams.atmosphereColor}
          transparent 
          opacity={planetParams.cloudDensity * 0.3}
        />
      </Sphere>
      
      <Points geometry={cloudGeometry}>
        <PointMaterial
          vertexColors
          size={0.8}
          sizeAttenuation={true}
          transparent
          opacity={0.6}
        />
      </Points>
      
      {/* Planetary rings */}
      {planetParams.hasRings && (
        <>
          <mesh rotation={[-Math.PI / 2, 0, 0]}>
            <ringGeometry args={[planetParams.size + 1, planetParams.size + 2, 64]} />
            <meshBasicMaterial 
              color="#888888" 
              transparent 
              opacity={0.4} 
              side={THREE.DoubleSide}
            />
          </mesh>
          <mesh rotation={[-Math.PI / 2, 0, 0]}>
            <ringGeometry args={[planetParams.size + 2.5, planetParams.size + 3, 64]} />
            <meshBasicMaterial 
              color="#666666" 
              transparent 
              opacity={0.3} 
              side={THREE.DoubleSide}
            />
          </mesh>
        </>
      )}
    </group>
  );
}