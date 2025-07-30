'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import { Box, Points, PointMaterial } from '@react-three/drei';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

interface Building {
  x: number;
  y: number;
  z: number;
  width: number;
  height: number;
  depth: number;
  color: THREE.Color;
  hasBeacon: boolean;
}

interface City3DProps {
  ipSeed: number;
}

export function City3D({ ipSeed }: City3DProps) {
  const groupRef = useRef<THREE.Group>(null);
  const beaconRefs = useRef<THREE.Mesh[]>([]);
  
  const cityParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    
    const numBuildings = 15 + Math.floor(rng.random() * 20); // 15-35 buildings
    const buildings: Building[] = [];
    const citySize = 20;
    
    // Generate buildings
    for (let i = 0; i < numBuildings; i++) {
      const x = (rng.random() - 0.5) * citySize;
      const z = (rng.random() - 0.5) * citySize;
      const height = 2 + rng.random() * 8; // 2-10 height
      const width = 1 + rng.random() * 2; // 1-3 width
      const depth = 1 + rng.random() * 2; // 1-3 depth
      
      // Building color based on height (taller = more modern = lighter)
      const heightRatio = height / 10;
      const hue = 0.15 + rng.random() * 0.1; // Yellowish to orange
      const saturation = 0.2 + heightRatio * 0.3;
      const lightness = 0.3 + heightRatio * 0.4;
      const color = new THREE.Color().setHSL(hue, saturation, lightness);
      
      buildings.push({
        x,
        y: height / 2, // Center Y at half height
        z,
        width,
        height,
        depth,
        color,
        hasBeacon: height > 6 && rng.random() > 0.7, // Tall buildings may have beacons
      });
    }
    
    return {
      numBuildings,
      buildings,
      timeOfDay: rng.random(), // 0 = night, 1 = day
      trafficDensity: 0.3 + rng.random() * 0.4, // 0.3-0.7
    };
  }, [ipSeed]);

  // Generate traffic/vehicle points
  const trafficGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 300);
    const numVehicles = Math.floor(cityParams.trafficDensity * 100);
    const positions = new Float32Array(numVehicles * 3);
    const colors = new Float32Array(numVehicles * 3);
    
    for (let i = 0; i < numVehicles; i++) {
      // Vehicles move along roads (grid pattern)
      const isVerticalRoad = rng.random() > 0.5;
      let x, z;
      
      if (isVerticalRoad) {
        x = (rng.random() - 0.5) * 18; // Moving along X axis
        z = Math.floor((rng.random() - 0.5) * 8) * 2; // Discrete road positions
      } else {
        x = Math.floor((rng.random() - 0.5) * 8) * 2; // Discrete road positions
        z = (rng.random() - 0.5) * 18; // Moving along Z axis
      }
      
      positions[i * 3] = x;
      positions[i * 3 + 1] = 0.2; // Slightly above ground
      positions[i * 3 + 2] = z;
      
      // Vehicle colors (red/white lights)
      const isRedLight = rng.random() > 0.5;
      if (isRedLight) {
        colors[i * 3] = 1.0; // Red
        colors[i * 3 + 1] = 0.2;
        colors[i * 3 + 2] = 0.2;
      } else {
        colors[i * 3] = 1.0; // White
        colors[i * 3 + 1] = 1.0;
        colors[i * 3 + 2] = 1.0;
      }
    }
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    
    return geometry;
  }, [ipSeed, cityParams.trafficDensity]);

  // Generate window lights
  const windowGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 400);
    const allWindows: number[] = [];
    const allColors: number[] = [];
    
    cityParams.buildings.forEach((building) => {
      const windowsPerFloor = Math.max(1, Math.floor(building.width * building.depth * 2));
      const floors = Math.max(1, Math.floor(building.height * 2));
      
      for (let floor = 0; floor < floors; floor++) {
        for (let window = 0; window < windowsPerFloor; window++) {
          if (rng.random() < 0.6) { // 60% of windows are lit
            const x = building.x + (rng.random() - 0.5) * building.width * 0.8;
            const y = building.y - building.height/2 + (floor + 0.5) * (building.height / floors);
            const z = building.z + (rng.random() - 0.5) * building.depth * 0.8;
            
            allWindows.push(x, y, z);
            
            // Window light color (warm yellow/orange)
            const brightness = 0.8 + rng.random() * 0.2;
            allColors.push(1.0 * brightness, 0.8 * brightness, 0.4 * brightness);
          }
        }
      }
    });
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(new Float32Array(allWindows), 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(new Float32Array(allColors), 3));
    
    return geometry;
  }, [ipSeed, cityParams.buildings]);

  useFrame((state) => {
    // Animate traffic movement
    if (trafficGeometry.attributes.position) {
      const positions = trafficGeometry.attributes.position.array as Float32Array;
      for (let i = 0; i < positions.length; i += 3) {
        // Simple traffic movement
        positions[i] += Math.sin(state.clock.elapsedTime + i) * 0.01;
        positions[i + 2] += Math.cos(state.clock.elapsedTime + i) * 0.01;
      }
      trafficGeometry.attributes.position.needsUpdate = true;
    }
    
    // Animate beacon lights
    beaconRefs.current.forEach((beacon, index) => {
      if (beacon) {
        const shouldBlink = Math.sin(state.clock.elapsedTime * 3 + index) > 0;
        beacon.visible = shouldBlink;
      }
    });
  });

  return (
    <group ref={groupRef}>
      {/* Ground plane */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.1, 0]}>
        <planeGeometry args={[25, 25]} />
        <meshLambertMaterial color="#333333" />
      </mesh>
      
      {/* Buildings */}
      {cityParams.buildings.map((building, index) => (
        <group key={index}>
          <Box
            args={[building.width, building.height, building.depth]}
            position={[building.x, building.y, building.z]}
          >
            <meshLambertMaterial color={building.color} />
          </Box>
          
          {/* Beacon light */}
          {building.hasBeacon && (
            <mesh
              ref={(ref) => {
                if (ref) beaconRefs.current[index] = ref;
              }}
              position={[building.x, building.y + building.height/2 + 0.2, building.z]}
            >
              <sphereGeometry args={[0.1, 8, 8]} />
              <meshBasicMaterial color="#ff0000" />
            </mesh>
          )}
        </group>
      ))}
      
      {/* Window lights */}
      <Points geometry={windowGeometry}>
        <PointMaterial
          vertexColors
          size={0.2}
          sizeAttenuation={true}
          transparent
          opacity={0.8}
        />
      </Points>
      
      {/* Traffic */}
      <Points geometry={trafficGeometry}>
        <PointMaterial
          vertexColors
          size={0.3}
          sizeAttenuation={true}
          transparent
          opacity={0.9}
        />
      </Points>
    </group>
  );
}