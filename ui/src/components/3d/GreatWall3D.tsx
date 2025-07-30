'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';
import { createWavePropagationShaderMaterial, GALAXY_SHADER_PRESETS } from '@/shaders/galaxyShaders';

interface GreatWall3DProps {
  ipSeed: number;
}

interface SplinePoint {
  position: THREE.Vector3;
  tangent: THREE.Vector3;
  width: number;
}

interface Galaxy {
  position: THREE.Vector3;
  distanceToSpline: number;
  size: number;
  filamentId: number;
  wavePosition: number; // Position along filament (0 to 1)
}

interface Filament {
  spline: SplinePoint[];
  galaxies: Galaxy[];
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

        // Calculate variable width along the spline with organic variation
        // Use position-based seed for consistent but varied base width
        const baseWidthSeed = t * 12345.67 + f * 9876.54;
        const baseWidth = 1.0 + (Math.sin(baseWidthSeed) * 0.5 + 0.5) * 2.0; // Base width 1-3

        // Multiple frequency variations for organic width changes
        // Use filament index and position for unique phase offsets
        const lowFreqVariation = Math.sin(t * Math.PI * 0.5 + f * 2.1 + s * 0.3) * 0.4;
        const midFreqVariation = Math.sin(t * Math.PI * 2 + f * 1.7 + s * 0.5) * 0.3;
        const highFreqVariation = Math.sin(t * Math.PI * 8 + f * 3.2 + s * 0.8) * 0.2;
        const noiseVariation = (Math.sin(t * 31.4 + f * 15.9 + s * 7.3) * 0.5) * 0.3;

        // Combine variations with different weights
        const combinedVariation = lowFreqVariation + midFreqVariation + highFreqVariation + noiseVariation;

        // Ensure minimum width and apply variation
        const width = baseWidth * (0.4 + 0.6 * (1 + combinedVariation));

        spline.push({ position: basePos, tangent, width });
      }

      // Create smooth curve from spline points for proper galaxy distribution
      const curvePoints = spline.map(point => point.position);
      const curve = new THREE.CatmullRomCurve3(curvePoints, false, 'centripetal');

      // Place many tiny galaxies along the smooth curve
      const galaxyCount = 5000 + Math.floor(rng.random() * 5000); // galaxies per filament
      const galaxies: Galaxy[] = [];

      for (let g = 0; g < galaxyCount; g++) {
        // Distribute galaxies more evenly along the curve with some clustering
        const clustering = rng.random();
        let t: number;

        if (clustering < 0.2) {
          // Some clustering near start
          t = rng.random() * 0.25;
        } else if (clustering < 0.4) {
          // Some clustering near end
          t = 0.75 + rng.random() * 0.25;
        } else {
          // Most galaxies distributed evenly
          t = rng.random();
        }

        // Get position along the smooth curve
        const curvePosition = curve.getPointAt(t);

        // Get tangent at this point for scatter direction
        const tangent = curve.getTangentAt(t).normalize();

        // Interpolate width smoothly between spline points
        const splineT = t * (spline.length - 1);
        const splineIndex = Math.floor(splineT);
        const localT = splineT - splineIndex;

        const currentPoint = spline[Math.min(splineIndex, spline.length - 1)];
        const nextPoint = spline[Math.min(splineIndex + 1, spline.length - 1)];

        // Linear interpolation between current and next point widths
        const filamentWidth = currentPoint.width + (nextPoint.width - currentPoint.width) * localT;

        // Scatter based on filament width
        const scatter1 = tangent.clone().cross(new THREE.Vector3(0, 1, 0)).normalize();
        const scatter2 = tangent.clone().cross(scatter1).normalize();

        const scatterAmount1 = (rng.random() - 0.5) * filamentWidth * 0.8;
        const scatterAmount2 = (rng.random() - 0.5) * filamentWidth * 0.8;

        const galaxyPos = curvePosition.clone();
        galaxyPos.add(scatter1.multiplyScalar(scatterAmount1));
        galaxyPos.add(scatter2.multiplyScalar(scatterAmount2));

        // Calculate distance to nearest spline point for glow intensity
        let minDistance = Infinity;
        for (const splinePoint of spline) {
          const distance = galaxyPos.distanceTo(splinePoint.position);
          minDistance = Math.min(minDistance, distance);
        }

        const galaxy: Galaxy = {
          position: galaxyPos,
          distanceToSpline: minDistance,
          size: 0.08 + rng.random() * 0.12,
          filamentId: f,
          wavePosition: t // Store the t parameter as wave position
        };

        galaxies.push(galaxy);
      }

      // Add isolated particle clusters at greater distances
      const clusterCount = 3 + Math.floor(rng.random() * 5); // 3-7 clusters per filament

      for (let c = 0; c < clusterCount; c++) {
        // Random position along the spline
        const clusterT = rng.random();
        const clusterPosition = curve.getPointAt(clusterT);

        // Get width at this position for reference
        const splineT = clusterT * (spline.length - 1);
        const splineIndex = Math.floor(splineT);
        const localT = splineT - splineIndex;
        const currentPoint = spline[Math.min(splineIndex, spline.length - 1)];
        const nextPoint = spline[Math.min(splineIndex + 1, spline.length - 1)];
        const localWidth = currentPoint.width + (nextPoint.width - currentPoint.width) * localT;

        // Place cluster at greater distance from spline
        const tangent = curve.getTangentAt(clusterT).normalize();
        const scatter1 = tangent.clone().cross(new THREE.Vector3(0, 1, 0)).normalize();
        const scatter2 = tangent.clone().cross(scatter1).normalize();

        // Cluster distance is 3-6x the local filament width
        const clusterDistance = localWidth * (3 + rng.random() * 3);
        const clusterDirection = scatter1.clone().multiplyScalar(rng.random() - 0.5)
          .add(scatter2.clone().multiplyScalar(rng.random() - 0.5)).normalize();

        const clusterCenter = clusterPosition.clone().add(clusterDirection.multiplyScalar(clusterDistance));

        // Generate particles within the cluster - fewer and irregular
        const particlesPerCluster = 15 + Math.floor(rng.random() * 25); // 15-40 particles (much fewer)
        const baseRadius = 0.3 + rng.random() * 0.8; // Smaller base size

        for (let p = 0; p < particlesPerCluster; p++) {
          // Irregular ellipsoidal distribution instead of perfect sphere
          const phi = rng.random() * Math.PI * 2;
          const cosTheta = rng.random() * 2 - 1;
          const sinTheta = Math.sqrt(1 - cosTheta * cosTheta);
          const r = Math.pow(rng.random(), 0.5) * baseRadius; // Less concentrated toward center

          // Create irregular shape with different radii in each axis
          const xStretch = 0.5 + rng.random() * 1.5; // 0.5x to 2x
          const yStretch = 0.5 + rng.random() * 1.5;
          const zStretch = 0.5 + rng.random() * 1.5;

          const particleOffset = new THREE.Vector3(
            r * sinTheta * Math.cos(phi) * xStretch,
            r * sinTheta * Math.sin(phi) * yStretch,
            r * cosTheta * zStretch
          );

          // Add some random noise for more irregular distribution
          const noise = new THREE.Vector3(
            (rng.random() - 0.5) * baseRadius * 0.3,
            (rng.random() - 0.5) * baseRadius * 0.3,
            (rng.random() - 0.5) * baseRadius * 0.3
          );

          particleOffset.add(noise);

          const particlePosition = clusterCenter.clone().add(particleOffset);

          // Calculate distance to nearest spline point for brightness
          let minDistance = Infinity;
          for (const splinePoint of spline) {
            const distance = particlePosition.distanceTo(splinePoint.position);
            minDistance = Math.min(minDistance, distance);
          }

          const clusterGalaxy: Galaxy = {
            position: particlePosition,
            distanceToSpline: minDistance,
            size: 0.04 + rng.random() * 0.08, // Smaller than main filament galaxies
            filamentId: f,
            wavePosition: clusterT // Use cluster position along filament
          };

          galaxies.push(clusterGalaxy);
        }
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

  // Create galaxy point data with distance-based brightness
  const galaxyPointData = useMemo(() => {
    const positions: number[] = [];
    const colors: number[] = [];
    const sizes: number[] = [];
    const wavePositions: number[] = [];
    const filamentIds: number[] = [];

    // Add galaxies to point data
    cosmicWeb.forEach((filament) => {
      filament.galaxies.forEach((galaxy) => {
        // Position
        positions.push(galaxy.position.x, galaxy.position.y, galaxy.position.z);

        // Calculate brightness based on distance to spline (closer = brighter)
        const maxDistance = 3.0; // Maximum expected distance for normalization
        const normalizedDistance = Math.min(galaxy.distanceToSpline / maxDistance, 1.0);
        const brightness = 1.0 - normalizedDistance * 0.7; // 0.3 to 1.0 range

        // Create color variation while maintaining filament preference
        const rng = new SeededRandom(ipSeed + galaxy.position.x * 1000 + galaxy.position.y * 1000 + galaxy.position.z * 1000);

        // Get filament color in HSL for easier manipulation
        const filamentHSL = { h: 0, s: 0, l: 0 };
        filament.color.getHSL(filamentHSL);

        // Create color variation
        const colorVariation = rng.random();
        let finalColor: THREE.Color;

        if (colorVariation < 0.5) {
          // 50% chance: stay close to filament color with small variations
          const hueShift = (rng.random() - 0.5) * 0.1; // ±18 degrees
          const satShift = (rng.random() - 0.5) * 0.3; // ±15% saturation
          const lightShift = (rng.random() - 0.5) * 0.2; // ±10% lightness

          finalColor = new THREE.Color().setHSL(
            (filamentHSL.h + hueShift + 1) % 1,
            Math.max(0, Math.min(1, filamentHSL.s + satShift)),
            Math.max(0, Math.min(1, filamentHSL.l + lightShift))
          );
        } else if (colorVariation < 0.9) {
          // 40% chance: complementary colors
          const complementaryHue = (filamentHSL.h + 0.5) % 1;
          const hueShift = (rng.random() - 0.5) * 0.2;

          finalColor = new THREE.Color().setHSL(
            (complementaryHue + hueShift + 1) % 1,
            filamentHSL.s * (0.7 + rng.random() * 0.3),
            filamentHSL.l * (0.8 + rng.random() * 0.4)
          );
        } else {
          // 10% chance: random colors for rare exotic galaxies
          finalColor = new THREE.Color().setHSL(
            rng.random(),
            0.5 + rng.random() * 0.5,
            0.3 + rng.random() * 0.4
          );
        }

        // Apply brightness to final color
        const brightColor = finalColor.multiplyScalar(brightness * 2.0);
        colors.push(brightColor.r, brightColor.g, brightColor.b);

        // Size based on brightness and galaxy size
        const pointSize = galaxy.size * (0.5 + brightness * 1.5); // Scale up for points
        sizes.push(pointSize);

        // Wave data for animation
        wavePositions.push(galaxy.wavePosition);
        filamentIds.push(galaxy.filamentId);
      });
    });

    return { positions, colors, sizes, wavePositions, filamentIds };
  }, [cosmicWeb]);

  // Create points mesh with custom glowing shader
  const galaxyPoints = useMemo(() => {
    if (galaxyPointData.positions.length === 0) return null;

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(galaxyPointData.positions, 3));
    geometry.setAttribute('color', new THREE.Float32BufferAttribute(galaxyPointData.colors, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(galaxyPointData.sizes, 1));
    geometry.setAttribute('wavePosition', new THREE.Float32BufferAttribute(galaxyPointData.wavePositions, 1));
    geometry.setAttribute('filamentId', new THREE.Float32BufferAttribute(galaxyPointData.filamentIds, 1));

    // Use wave propagation shader with cosmic web preset
    const material = createWavePropagationShaderMaterial(cosmicWeb.length, GALAXY_SHADER_PRESETS.cosmicWeb);

    return new THREE.Points(geometry, material);
  }, [galaxyPointData]);

  // Animation
  useFrame((state) => {
    // Update shader uniforms for wave propagation and pulsing effect
    if (galaxyPoints && galaxyPoints.material instanceof THREE.ShaderMaterial) {
      galaxyPoints.material.uniforms.time.value = state.clock.elapsedTime;
    }
  });

  return (
    <group ref={groupRef}>
      {/* Glowing galaxy points forming the cosmic web structure */}
      {galaxyPoints && (
        <primitive object={galaxyPoints} />
      )}
    </group>
  );
}