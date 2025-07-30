'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import { Sphere, Ring, Points, PointMaterial } from '@react-three/drei';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

interface Planet {
  size: number;
  orbitDistance: number;
  orbitSpeed: number;
  angle: number;
  moons: number;
  hasRings: boolean;
  type: 'rocky' | 'gas' | 'ice';
  color: THREE.Color;
}

interface SolarSystem3DProps {
  ipSeed: number;
}

export function SolarSystem3D({ ipSeed }: SolarSystem3DProps) {
  const groupRef = useRef<THREE.Group>(null);
  const planetRefs = useRef<THREE.Group[]>([]);
  
  // Generate solar system parameters
  const systemParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    
    const numPlanets = 4 + Math.floor(rng.random() * 8); // 4-12 planets
    const planets: Planet[] = [];
    
    // Generate planets
    for (let i = 0; i < numPlanets; i++) {
      const typeRoll = rng.random();
      let type: 'rocky' | 'gas' | 'ice';
      let baseSize: number;
      let color: THREE.Color;
      
      if (typeRoll < 0.5) {
        type = 'rocky';
        baseSize = 0.4 + rng.random() * 0.8; // 0.4 to 1.2
        color = new THREE.Color().setHSL(0.08 + rng.random() * 0.05, 0.7, 0.6); // Orange-brown
      } else if (typeRoll < 0.8) {
        type = 'gas';
        baseSize = 1.5 + rng.random() * 1.0; // 1.5 to 2.5
        color = new THREE.Color().setHSL(0.15 + rng.random() * 0.05, 0.6, 0.8); // Yellow
      } else {
        type = 'ice';
        baseSize = 1.0 + rng.random() * 0.8; // 1.0 to 1.8
        color = new THREE.Color().setHSL(0.55 + rng.random() * 0.1, 0.7, 0.7); // Light blue
      }
      
      const orbitDistance = 3 + i * 2.5 + rng.random() * 1.5;
      const orbitSpeed = 0.5 / Math.pow(orbitDistance / 3, 1.5); // Kepler's laws approximation
      
      let maxMoons = Math.floor(baseSize * 3);
      if (type === 'rocky') maxMoons = Math.min(maxMoons, 2);
      const moons = Math.floor(rng.random() * (maxMoons + 1));
      
      const hasRings = type === 'gas' ? rng.random() > 0.5 : 
                      type === 'ice' ? rng.random() > 0.8 : false;
      
      planets.push({
        size: baseSize,
        orbitDistance,
        orbitSpeed,
        angle: rng.random() * Math.PI * 2,
        moons,
        hasRings,
        type,
        color
      });
    }
    
    return {
      numPlanets,
      planets,
      starColor: new THREE.Color().setHSL(0.15 + rng.random() * 0.05, 0.8, 0.9), // Yellow-white
      asteroidBelt: {
        innerRadius: planets.length > 3 ? planets[3].orbitDistance + 1 : 10,
        outerRadius: planets.length > 3 ? planets[3].orbitDistance + 3 : 12,
        particles: 200
      }
    };
  }, [ipSeed]);

  // Generate asteroid belt
  const asteroidGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 300);
    const positions = new Float32Array(systemParams.asteroidBelt.particles * 3);
    const colors = new Float32Array(systemParams.asteroidBelt.particles * 3);
    
    for (let i = 0; i < systemParams.asteroidBelt.particles; i++) {
      const radius = systemParams.asteroidBelt.innerRadius + 
                    rng.random() * (systemParams.asteroidBelt.outerRadius - systemParams.asteroidBelt.innerRadius);
      const theta = rng.random() * Math.PI * 2;
      const y = (rng.random() - 0.5) * 0.5; // Thin disk
      
      positions[i * 3] = radius * Math.cos(theta);
      positions[i * 3 + 1] = y;
      positions[i * 3 + 2] = radius * Math.sin(theta);
      
      const grayValue = 0.3 + rng.random() * 0.3;
      colors[i * 3] = grayValue;
      colors[i * 3 + 1] = grayValue;
      colors[i * 3 + 2] = grayValue;
    }
    
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    
    return geometry;
  }, [ipSeed, systemParams.asteroidBelt]);

  // Animation loop
  useFrame((state, delta) => {
    systemParams.planets.forEach((planet, index) => {
      if (planetRefs.current[index]) {
        // Update orbital position
        planet.angle += planet.orbitSpeed * delta;
        
        const x = planet.orbitDistance * Math.cos(planet.angle);
        const z = planet.orbitDistance * Math.sin(planet.angle);
        
        planetRefs.current[index].position.set(x, 0, z);
        
        // Rotate planet
        planetRefs.current[index].rotation.y += delta * 2;
      }
    });
  });

  return (
    <group ref={groupRef}>
      {/* Central star */}
      <Sphere args={[1.5, 32, 32]} position={[0, 0, 0]}>
        <meshBasicMaterial color={systemParams.starColor} />
      </Sphere>
      
      {/* Star glow effect */}
      <Sphere args={[2.5, 16, 16]} position={[0, 0, 0]}>
        <meshBasicMaterial 
          color={systemParams.starColor} 
          transparent 
          opacity={0.3}
        />
      </Sphere>

      {/* Orbit paths */}
      {systemParams.planets.map((planet, index) => (
        <Ring
          key={`orbit-${index}`}
          args={[planet.orbitDistance - 0.05, planet.orbitDistance + 0.05, 64]}
          rotation={[-Math.PI / 2, 0, 0]}
        >
          <meshBasicMaterial color="#333333" transparent opacity={0.3} />
        </Ring>
      ))}

      {/* Planets */}
      {systemParams.planets.map((planet, index) => (
        <group
          key={`planet-${index}`}
          ref={(ref) => {
            if (ref) planetRefs.current[index] = ref;
          }}
        >
          {/* Planet sphere */}
          <Sphere args={[planet.size, 16, 16]}>
            <meshLambertMaterial color={planet.color} />
          </Sphere>
          
          {/* Planet rings */}
          {planet.hasRings && (
            <Ring
              args={[planet.size + 0.5, planet.size + 1.5, 32]}
              rotation={[-Math.PI / 2 + Math.PI / 8, 0, 0]}
            >
              <meshBasicMaterial 
                color="#888888" 
                transparent 
                opacity={0.6} 
                side={THREE.DoubleSide}
              />
            </Ring>
          )}
          
          {/* Moons */}
          {Array.from({ length: planet.moons }, (_, moonIndex) => {
            const moonDistance = planet.size + 2 + moonIndex * 1.5;
            const moonAngle = (moonIndex * Math.PI * 2) / planet.moons;
            
            return (
              <Sphere
                key={`moon-${moonIndex}`}
                args={[0.2, 8, 8]}
                position={[
                  moonDistance * Math.cos(moonAngle),
                  0,
                  moonDistance * Math.sin(moonAngle)
                ]}
              >
                <meshLambertMaterial color="#cccccc" />
              </Sphere>
            );
          })}
        </group>
      ))}

      {/* Asteroid belt */}
      <Points geometry={asteroidGeometry}>
        <PointMaterial
          vertexColors
          size={0.3}
          sizeAttenuation={true}
          transparent
          opacity={0.7}
        />
      </Points>
    </group>
  );
}