'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

interface GreatWall3DProps {
  ipSeed: number;
}

interface SplinePoint {
  position: THREE.Vector3;
  tangent: THREE.Vector3;
}

interface Filament {
  spline: SplinePoint[];
  galaxies: THREE.Vector3[];
  color: THREE.Color;
}

export function GreatWall3D({ ipSeed }: GreatWall3DProps) {
  const groupRef = useRef<THREE.Group>(null);

  // Generate cosmic web structure
  const cosmicWeb = useMemo(() => {
    const rng = new SeededRandom(ipSeed || 1); // Ensure non-zero seed
    const numFilaments = 5 + Math.floor(rng.random() * 4); // 5-8 filaments
    const filaments: Filament[] = [];

    // Generate main anchor points for the cosmic web
    const anchorPoints: THREE.Vector3[] = [];
    for (let i = 0; i < 8; i++) {
      anchorPoints.push(new THREE.Vector3(
        (rng.random() - 0.5) * 40,
        (rng.random() - 0.5) * 30,
        (rng.random() - 0.5) * 20
      ));
    }

    // Create filaments connecting anchor points
    for (let f = 0; f < numFilaments; f++) {
      let startAnchor = anchorPoints[rng.randomInt(0, anchorPoints.length)];
      let endAnchor = anchorPoints[rng.randomInt(0, anchorPoints.length)];

      // Ensure we don't connect an anchor to itself
      let attempts = 0;
      while (startAnchor === endAnchor && attempts < 10) {
        endAnchor = anchorPoints[rng.randomInt(0, anchorPoints.length)];
        attempts++;
      }

      if (startAnchor === endAnchor) {
        continue;
      }

      // Generate spline points between anchors
      const splineResolution = 20 + Math.floor(rng.random() * 20); // 20-40 points
      const spline: SplinePoint[] = [];

      for (let s = 0; s <= splineResolution; s++) {
        const t = s / splineResolution;

        // Base interpolation between anchor points
        const basePos = startAnchor.clone().lerp(endAnchor, t);

        // Add organic waviness with multiple frequencies
        const waveAmplitude = 3 + rng.random() * 5;
        const wave1 = Math.sin(t * Math.PI * 2 + f * 0.5) * waveAmplitude;
        const wave2 = Math.sin(t * Math.PI * 6 + f * 1.2) * waveAmplitude * 0.3;
        const wave3 = Math.sin(t * Math.PI * 12 + f * 2.1) * waveAmplitude * 0.1;

        // Apply waves in perpendicular directions
        const direction = endAnchor.clone().sub(startAnchor).normalize();
        const perpendicular1 = new THREE.Vector3(-direction.y, direction.x, 0).normalize();
        const perpendicular2 = direction.clone().cross(perpendicular1).normalize();

        basePos.add(perpendicular1.clone().multiplyScalar(wave1 + wave2));
        basePos.add(perpendicular2.clone().multiplyScalar(wave3));

        // Add some random noise for organic feel
        basePos.add(new THREE.Vector3(
          (rng.random() - 0.5) * 2,
          (rng.random() - 0.5) * 2,
          (rng.random() - 0.5) * 2
        ));

        // Calculate tangent for smooth curves
        let tangent: THREE.Vector3;
        if (s === 0) {
          tangent = direction.clone();
        } else if (s === splineResolution) {
          tangent = direction.clone();
        } else {
          // Approximate tangent from neighboring points
          const prevT = (s - 1) / splineResolution;
          const nextT = (s + 1) / splineResolution;
          const prevPos = startAnchor.clone().lerp(endAnchor, prevT);
          const nextPos = startAnchor.clone().lerp(endAnchor, nextT);
          tangent = nextPos.sub(prevPos).normalize();
        }

        spline.push({ position: basePos, tangent });
      }

      // Place galaxies along the spline
      const galaxyCount = 15 + Math.floor(rng.random() * 25); // 15-40 galaxies per filament
      const galaxies: THREE.Vector3[] = [];

      for (let g = 0; g < galaxyCount; g++) {
        // Distribute galaxies non-uniformly (cluster more at certain points)
        const clustering = rng.random();
        let t: number;

        if (clustering < 0.3) {
          // Cluster near start
          t = rng.random() * 0.3;
        } else if (clustering < 0.6) {
          // Cluster near end
          t = 0.7 + rng.random() * 0.3;
        } else {
          // Distribute randomly
          t = rng.random();
        }

        const splineIndex = Math.floor(t * (spline.length - 1));
        const splinePoint = spline[splineIndex];

        // Add some scatter around the spline
        const scatter = splinePoint.tangent.clone().cross(new THREE.Vector3(0, 1, 0)).normalize();
        const scatterAmount = (rng.random() - 0.5) * 2;

        const galaxyPos = splinePoint.position.clone();
        galaxyPos.add(scatter.multiplyScalar(scatterAmount));

        galaxies.push(galaxyPos);
      }

      // Assign color based on filament properties
      const hue = 0.55 + rng.random() * 0.15; // Blue to cyan range
      const saturation = 0.6 + rng.random() * 0.3;
      const lightness = 0.4 + rng.random() * 0.4;
      const color = new THREE.Color().setHSL(hue, saturation, lightness);

      filaments.push({ spline, galaxies, color });
    }

    return filaments;
  }, [ipSeed]);

  // Create render geometries
  const { splineGeometries, galaxyInstanceData } = useMemo(() => {
    const splineGeometries: THREE.BufferGeometry[] = [];
    const galaxyInstanceData: { positions: THREE.Vector3[], colors: THREE.Color[], scales: number[] } = {
      positions: [],
      colors: [],
      scales: []
    };

    // Generate smooth spline geometries using CatmullRomCurve3
    cosmicWeb.forEach((filament) => {
      if (filament.spline.length < 3) return; // Need at least 3 points for smooth curve

      // Extract positions for curve
      const curvePoints = filament.spline.map(point => point.position);

      // Create smooth curve
      const curve = new THREE.CatmullRomCurve3(curvePoints, false, 'centripetal');

      // Sample points along the curve for higher resolution
      const curveResolution = 100; // Higher resolution for smoother curves
      const curvePositions = curve.getPoints(curveResolution);

      // Convert to line segments for rendering
      const segments = curvePositions.length - 1;
      const positions = new Float32Array(segments * 6);
      const colors = new Float32Array(segments * 6);

      for (let i = 0; i < segments; i++) {
        const point1 = curvePositions[i];
        const point2 = curvePositions[i + 1];

        // First point of segment
        positions[i * 6] = point1.x;
        positions[i * 6 + 1] = point1.y;
        positions[i * 6 + 2] = point1.z;

        // Second point of segment
        positions[i * 6 + 3] = point2.x;
        positions[i * 6 + 4] = point2.y;
        positions[i * 6 + 5] = point2.z;

        // Colors for both points
        colors[i * 6] = filament.color.r * 0.3;
        colors[i * 6 + 1] = filament.color.g * 0.3;
        colors[i * 6 + 2] = filament.color.b * 0.3;
        colors[i * 6 + 3] = filament.color.r * 0.3;
        colors[i * 6 + 4] = filament.color.g * 0.3;
        colors[i * 6 + 5] = filament.color.b * 0.3;
      }

      const geometry = new THREE.BufferGeometry();
      geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
      geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
      splineGeometries.push(geometry);

      // Add galaxies to instance data
      filament.galaxies.forEach((galaxy) => {
        galaxyInstanceData.positions.push(galaxy.clone());
        galaxyInstanceData.colors.push(filament.color.clone());
        // Vary galaxy sizes based on seeded random
        const rng = new SeededRandom(ipSeed + galaxy.x + galaxy.y + galaxy.z);
        const scale = 0.5 + rng.random() * 1.5; // 0.5 to 2.0 scale
        galaxyInstanceData.scales.push(scale);
      });
    });

    return { splineGeometries, galaxyInstanceData };
  }, [cosmicWeb, ipSeed]);

  // Create instancedMesh for galaxies
  const galaxyInstancedMesh = useMemo(() => {
    if (galaxyInstanceData.positions.length === 0) return null;

    const sphereGeometry = new THREE.SphereGeometry(0.8, 8, 6); // Low-poly sphere for performance
    const material = new THREE.MeshBasicMaterial({ 
      transparent: true, 
      opacity: 0.8,
      blending: THREE.AdditiveBlending
    });
    
    const instancedMesh = new THREE.InstancedMesh(sphereGeometry, material, galaxyInstanceData.positions.length);
    
    // Set up instance matrices and colors
    const matrix = new THREE.Matrix4();
    const color = new THREE.Color();
    
    for (let i = 0; i < galaxyInstanceData.positions.length; i++) {
      const position = galaxyInstanceData.positions[i];
      const instanceColor = galaxyInstanceData.colors[i];
      const scale = galaxyInstanceData.scales[i];
      
      // Set position and scale
      matrix.makeScale(scale, scale, scale);
      matrix.setPosition(position);
      instancedMesh.setMatrixAt(i, matrix);
      
      // Set color
      color.copy(instanceColor);
      instancedMesh.setColorAt(i, color);
    }
    
    instancedMesh.instanceMatrix.needsUpdate = true;
    if (instancedMesh.instanceColor) {
      instancedMesh.instanceColor.needsUpdate = true;
    }
    
    return instancedMesh;
  }, [galaxyInstanceData]);

  // Animation
  useFrame((state) => {
    if (groupRef.current) {
      groupRef.current.rotation.y = state.clock.elapsedTime * 0.02;
      groupRef.current.rotation.x = Math.sin(state.clock.elapsedTime * 0.1) * 0.1;
    }
  });

  return (
    <group ref={groupRef}>
      {/* Filament splines */}
      {splineGeometries.map((geometry, index) => (
        <lineSegments key={`spline-${index}`} geometry={geometry}>
          <lineBasicMaterial vertexColors />
        </lineSegments>
      ))}

      {/* Galaxies as spheres */}
      {galaxyInstancedMesh && (
        <primitive object={galaxyInstancedMesh} />
      )}
    </group>
  );
}