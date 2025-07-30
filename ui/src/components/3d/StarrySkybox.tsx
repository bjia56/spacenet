'use client';

import { useMemo } from 'react';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

interface StarrySkyboxProps {
  seed?: number;
  starCount?: number;
  skyboxRadius?: number;
}

export function StarrySkybox({
  seed = 42,
  starCount = 2000,
  skyboxRadius = 500
}: StarrySkyboxProps) {

  const starGeometry = useMemo(() => {
    const rng = new SeededRandom(seed);
    const positions: number[] = [];
    const colors: number[] = [];
    const sizes: number[] = [];

    for (let i = 0; i < starCount; i++) {
      // Generate stars on a sphere surface
      const theta = rng.random() * Math.PI * 2; // Azimuth
      const phi = Math.acos(2 * rng.random() - 1); // Polar angle for uniform distribution

      const x = skyboxRadius * Math.sin(phi) * Math.cos(theta);
      const y = skyboxRadius * Math.sin(phi) * Math.sin(theta);
      const z = skyboxRadius * Math.cos(phi);

      positions.push(x, y, z);

      // Star color variations - from blue-white to yellow-white
      const starType = rng.random();
      let r, g, b;

      if (starType < 0.1) {
        // Blue giants (rare)
        r = 0.7 + rng.random() * 0.3;
        g = 0.8 + rng.random() * 0.2;
        b = 1.0;
      } else if (starType < 0.3) {
        // White stars
        const intensity = 0.8 + rng.random() * 0.2;
        r = g = b = intensity;
      } else if (starType < 0.7) {
        // Yellow-white stars (like our sun)
        r = 1.0;
        g = 0.9 + rng.random() * 0.1;
        b = 0.7 + rng.random() * 0.2;
      } else if (starType < 0.9) {
        // Orange stars
        r = 1.0;
        g = 0.6 + rng.random() * 0.3;
        b = 0.3 + rng.random() * 0.3;
      } else {
        // Red giants (rare)
        r = 1.0;
        g = 0.3 + rng.random() * 0.3;
        b = 0.1 + rng.random() * 0.2;
      }

      colors.push(r, g, b);

      // Varying star sizes based on brightness/distance illusion
      const brightness = rng.random();
      let size;
      if (brightness < 0.7) {
        // Most stars are small
        size = 0.5 + rng.random() * 1.0;
      } else if (brightness < 0.9) {
        // Some medium stars
        size = 1.5 + rng.random() * 1.5;
      } else {
        // Few bright stars
        size = 3.0 + rng.random() * 2.0;
      }
      sizes.push(size);
    }

    // Create nebula effect with some larger, dimmer points
    const nebulaCount = Math.floor(starCount * 0.05); // 5% nebula points
    for (let i = 0; i < nebulaCount; i++) {
      const theta = rng.random() * Math.PI * 2;
      const phi = Math.acos(2 * rng.random() - 1);

      const x = skyboxRadius * 0.98 * Math.sin(phi) * Math.cos(theta);
      const y = skyboxRadius * 0.98 * Math.sin(phi) * Math.sin(theta);
      const z = skyboxRadius * 0.98 * Math.cos(phi);

      positions.push(x, y, z);

      // Nebula colors - purples, magentas, and deep blues
      const nebulaType = rng.random();
      let r, g, b;

      if (nebulaType < 0.4) {
        // Purple nebula
        r = 0.4 + rng.random() * 0.4;
        g = 0.1 + rng.random() * 0.3;
        b = 0.6 + rng.random() * 0.4;
      } else if (nebulaType < 0.7) {
        // Magenta nebula
        r = 0.5 + rng.random() * 0.4;
        g = 0.1 + rng.random() * 0.2;
        b = 0.4 + rng.random() * 0.4;
      } else {
        // Deep blue nebula
        r = 0.1 + rng.random() * 0.2;
        g = 0.2 + rng.random() * 0.3;
        b = 0.5 + rng.random() * 0.5;
      }

      colors.push(r, g, b);
      sizes.push(8.0 + rng.random() * 12.0); // Large, diffuse points
    }

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.Float32BufferAttribute(colors, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(sizes, 1));

    return geometry;
  }, [seed, starCount, skyboxRadius]);

  return (
    <points geometry={starGeometry}>
      <pointsMaterial
        vertexColors
        size={1}
        sizeAttenuation={true} // Keep stars same size regardless of distance
        transparent
        opacity={0.8}
        blending={THREE.AdditiveBlending} // Makes stars glow
      />
    </points>
  );
}