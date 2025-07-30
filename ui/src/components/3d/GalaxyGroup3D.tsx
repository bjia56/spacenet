'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';
import { createGalaxyShaderMaterial, GALAXY_SHADER_PRESETS } from '@/shaders/galaxyShaders';

interface GalaxyGroup3DProps {
  ipSeed: number;
}

export function GalaxyGroup3D({ ipSeed }: GalaxyGroup3DProps) {
  const groupRef = useRef<THREE.Group>(null);

  const groupParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);

    return {
      numGalaxies: 15 + Math.floor(rng.random() * 20), // 15-35 galaxies
      streamLength: 25 + rng.random() * 15, // Vary stream length 25-40
      streamCurvature: 0.5 + rng.random() * 1.5, // Stream curvature intensity
      streamTwist: 0.3 + rng.random() * 0.8, // Stream twist amount
      branchiness: rng.random() * 0.6, // How much the stream branches/spreads
      colors: {
        bright: new THREE.Color().setHSL(0.6 + rng.random() * 0.1, 0.9, 0.8),
        medium: new THREE.Color().setHSL(0.55 + rng.random() * 0.1, 0.7, 0.6),
        dim: new THREE.Color().setHSL(0.65 + rng.random() * 0.1, 0.5, 0.4),
      }
    };
  }, [ipSeed]);

  const streamData = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    const totalPoints = 3500 + Math.floor(rng.random() * 1500); // 3500-5000 points for richer appearance
    const positions = new Float32Array(totalPoints * 3);
    const colors = new Float32Array(totalPoints * 3);
    const sizes = new Float32Array(totalPoints);

    // GPU rotation attributes
    const originalPositions = new Float32Array(totalPoints * 3);
    const rotationCenters = new Float32Array(totalPoints * 3);
    const rotationAxes = new Float32Array(totalPoints * 3);
    const rotationSpeeds = new Float32Array(totalPoints);

    // Dual rotation attributes for orbital motion
    const groupCenters = new Float32Array(totalPoints * 3);
    const orbitalSpeeds = new Float32Array(totalPoints);

    // Create individual galaxy centers along the stream
    const numGalaxies = groupParams.numGalaxies;
    const galaxyCenters: THREE.Vector3[] = [];
    const galaxyRadii: number[] = [];

    for (let g = 0; g < numGalaxies; g++) {
      const galaxyT = (g + 0.5) / numGalaxies; // Evenly distribute along stream

      // Add more complex path variation with multiple harmonics
      const primaryFlow = Math.sin(galaxyT * Math.PI * 3 + rng.random() * Math.PI) * 3;
      const secondaryFlow = Math.sin(galaxyT * Math.PI * 7 + rng.random() * Math.PI) * 1.5;
      const tertiaryFlow = Math.sin(galaxyT * Math.PI * 11 + rng.random() * Math.PI) * 0.8;
      const totalFlow = primaryFlow + secondaryFlow + tertiaryFlow;

      // More varied stream path with curvature using group parameters
      const streamCurvature = Math.sin(galaxyT * Math.PI * 2.3) * 4 * groupParams.streamCurvature;
      const streamTwist = Math.cos(galaxyT * Math.PI * 1.7) * 2 * groupParams.streamTwist;

      // Add branching effect
      const branchOffset = (rng.random() - 0.5) * groupParams.branchiness * 6;

      const centerX = (galaxyT - 0.5) * groupParams.streamLength + totalFlow + streamCurvature + branchOffset;
      const centerY = Math.sin(galaxyT * Math.PI * 5 + rng.random() * Math.PI) * 2.5 * groupParams.streamTwist + streamTwist;
      const centerZ = Math.cos(galaxyT * Math.PI * 4.2 + rng.random() * Math.PI) * 2.2 * groupParams.streamCurvature + Math.sin(galaxyT * Math.PI * 8) * 1.2;

      galaxyCenters.push(new THREE.Vector3(centerX, centerY, centerZ));
      galaxyRadii.push(1.5 + rng.random() * 4); // Galaxy radius 1.5-5.5 units for more variety
    }

    // Group center for orbital motion (center of all galaxies)
    const groupCenter = new THREE.Vector3(0, 0, 0);

    for (let i = 0; i < totalPoints; i++) {
      // Assign point to nearest galaxy
      const t = i / totalPoints;
      let nearestGalaxy = 0;
      let minDistance = Infinity;

      // Find which galaxy this point belongs to based on stream position
      const streamT = t;
      for (let g = 0; g < numGalaxies; g++) {
        const galaxyT = (g + 0.5) / numGalaxies;
        const distance = Math.abs(streamT - galaxyT);
        if (distance < minDistance) {
          minDistance = distance;
          nearestGalaxy = g;
        }
      }

      const galaxyCenter = galaxyCenters[nearestGalaxy];
      const galaxyRadius = galaxyRadii[nearestGalaxy];

      // Generate point within galaxy's spiral arms
      const spiralRng = new SeededRandom(ipSeed + nearestGalaxy * 2000 + i);
      
      // Spiral galaxy parameters
      const numArms = 2 + Math.floor(spiralRng.random() * 3); // 2-4 spiral arms
      const spiralTightness = 0.3 + spiralRng.random() * 0.5; // How tight the spiral is
      const armWidth = 0.4 + spiralRng.random() * 0.3; // Width of spiral arms
      
      // Radial distance (concentrated toward center)
      const localRadius = Math.pow(rng.random(), 2) * galaxyRadius;
      
      // Choose which spiral arm this point belongs to
      const armIndex = Math.floor(rng.random() * numArms);
      const armAngleOffset = (armIndex * 2 * Math.PI) / numArms;
      
      // Spiral angle based on radius (tighter spirals further out)
      const spiralAngle = localRadius * spiralTightness + armAngleOffset;
      
      // Add some randomness within the arm width
      const armDeviation = (rng.random() - 0.5) * armWidth;
      const finalAngle = spiralAngle + armDeviation;
      
      // Also add some random points for galaxy halo/bulge (30% chance)
      const isHaloPoint = rng.random() < 0.3;
      const localTheta = isHaloPoint ? 
        rng.random() * Math.PI * 2 : // Random angle for halo
        finalAngle; // Spiral arm angle
      
      // Flatten the galaxy (disk-like structure)
      const diskThickness = 0.15 + spiralRng.random() * 0.1; // Galaxy disk thickness
      const localZ = (rng.random() - 0.5) * galaxyRadius * diskThickness;
      
      const localX = localRadius * Math.cos(localTheta);
      const localY = localRadius * Math.sin(localTheta);

      const x = galaxyCenter.x + localX;
      const y = galaxyCenter.y + localY;
      const z = galaxyCenter.z + localZ;

      positions[i * 3] = x;
      positions[i * 3 + 1] = y;
      positions[i * 3 + 2] = z;

      // Store original positions for GPU rotation
      originalPositions[i * 3] = x;
      originalPositions[i * 3 + 1] = y;
      originalPositions[i * 3 + 2] = z;

      // Set rotation center to the galaxy center (not stream center)
      rotationCenters[i * 3] = galaxyCenter.x;
      rotationCenters[i * 3 + 1] = galaxyCenter.y;
      rotationCenters[i * 3 + 2] = galaxyCenter.z;

      // Random rotation axis for each galaxy for variety
      const galaxyAxisRng = new SeededRandom(ipSeed + nearestGalaxy * 1000);
      const rotationAxis = new THREE.Vector3(
        galaxyAxisRng.random() - 0.5,
        0.5 + galaxyAxisRng.random() * 0.5, // Prefer Y-axis but add variation
        galaxyAxisRng.random() - 0.5
      ).normalize();

      rotationAxes[i * 3] = rotationAxis.x;
      rotationAxes[i * 3 + 1] = rotationAxis.y;
      rotationAxes[i * 3 + 2] = rotationAxis.z;

      // Individual galaxy rotation speed (slower and more realistic)
      const individualSpeed = (0.5 / galaxyRadius) * (0.8 + rng.random() * 0.4);
      rotationSpeeds[i] = individualSpeed;

      // Group center for orbital motion
      groupCenters[i * 3] = groupCenter.x;
      groupCenters[i * 3 + 1] = groupCenter.y;
      groupCenters[i * 3 + 2] = groupCenter.z;

      // Orbital speed based on distance from group center (closer = faster, like Kepler's laws)
      const distanceFromGroupCenter = galaxyCenter.distanceTo(groupCenter);
      const maxDistance = groupParams.streamLength * 0.8; // Approximate max distance
      const normalizedDistance = Math.min(distanceFromGroupCenter / maxDistance, 1.0);
      
      // Inverse relationship: closer galaxies orbit faster (like planetary motion)
      const baseOrbitalSpeed = 0.04; // Base orbital speed
      const orbitalSpeed = baseOrbitalSpeed * (1.0 - normalizedDistance * 0.7); // 30%-100% of base speed
      orbitalSpeeds[i] = orbitalSpeed;

      // Size variation based on galaxy structure
      const distanceFromGalaxyCenter = localRadius;
      const normalizedRadius = distanceFromGalaxyCenter / galaxyRadius;
      
      // Brighter/larger points in spiral arms and galaxy center
      let galaxySize;
      if (isHaloPoint) {
        // Halo points are dimmer
        galaxySize = 0.2 + rng.random() * 0.3;
      } else {
        // Spiral arm points are brighter, especially in the central region
        const armBrightness = normalizedRadius < 0.3 ? 
          0.7 + rng.random() * 0.4 : // Bright galactic center
          0.4 + rng.random() * 0.5;   // Moderate spiral arms
        galaxySize = armBrightness;
      }
      sizes[i] = galaxySize;

      // Color based on galaxy structure with intra-galaxy variation
      let baseColor;
      if (t < 0.2 || t > 0.8) {
        baseColor = groupParams.colors.dim;
      } else if (t < 0.4 || t > 0.6) {
        baseColor = groupParams.colors.medium;
      } else {
        baseColor = groupParams.colors.bright;
      }
      
      // Add galaxy-specific color variations
      const galaxyHue = (nearestGalaxy * 0.1) % 1.0; // Each galaxy gets a slight hue shift
      const spiralColorRng = new SeededRandom(ipSeed + nearestGalaxy * 3000 + Math.floor(i / 10));
      
      let finalColor = baseColor.clone();
      
      // Different colors for different galaxy regions
      if (isHaloPoint) {
        // Halo: Older, redder stars
        const hsl = { h: 0, s: 0, l: 0 };
        baseColor.getHSL(hsl);
        finalColor.setHSL(
          (hsl.h + galaxyHue + 0.05 + spiralColorRng.random() * 0.1) % 1, // Slightly redder
          Math.max(0.2, hsl.s - 0.2 + spiralColorRng.random() * 0.2), // Less saturated
          Math.max(0.2, hsl.l - 0.1 + spiralColorRng.random() * 0.2)  // Dimmer
        );
      } else {
        // Spiral arms: Mix of young blue stars and older populations
        const hsl = { h: 0, s: 0, l: 0 };
        baseColor.getHSL(hsl);
        
        const colorVariation = spiralColorRng.random();
        if (colorVariation < 0.3) {
          // 30% chance: Young blue/white stars in spiral arms
          finalColor.setHSL(
            (hsl.h + galaxyHue - 0.15 + spiralColorRng.random() * 0.1) % 1, // Bluer
            Math.min(1.0, hsl.s + 0.1 + spiralColorRng.random() * 0.2), // More saturated
            Math.min(1.0, hsl.l + 0.1 + spiralColorRng.random() * 0.2)  // Brighter
          );
        } else if (colorVariation < 0.6) {
          // 30% chance: Star forming regions (pinkish/red emission)
          finalColor.setHSL(
            (hsl.h + galaxyHue + 0.8 + spiralColorRng.random() * 0.1) % 1, // Pinkish-red
            Math.min(1.0, hsl.s + 0.2 + spiralColorRng.random() * 0.3), // Highly saturated
            Math.max(0.3, hsl.l + spiralColorRng.random() * 0.2)        // Variable brightness
          );
        } else {
          // 40% chance: Mixed stellar populations (slight variation of base)
          finalColor.setHSL(
            (hsl.h + galaxyHue + (spiralColorRng.random() - 0.5) * 0.1) % 1,
            Math.max(0.3, hsl.s + (spiralColorRng.random() - 0.5) * 0.3),
            Math.max(0.2, hsl.l + (spiralColorRng.random() - 0.5) * 0.3)
          );
        }
      }
      
      const color = finalColor;

      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;
    }

    return {
      positions,
      colors,
      sizes,
      originalPositions,
      rotationCenters,
      rotationAxes,
      rotationSpeeds,
      groupCenters,
      orbitalSpeeds
    };
  }, [ipSeed, groupParams]);

  // Create geometry and material with GPU dual rotation
  const streamPoints = useMemo(() => {
    if (!streamData.positions.length) return null;

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(streamData.positions, 3));
    geometry.setAttribute('color', new THREE.Float32BufferAttribute(streamData.colors, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(streamData.sizes, 1));

    // Add rotation attributes for GPU-based dual rotation (individual + orbital)
    geometry.setAttribute('originalPosition', new THREE.Float32BufferAttribute(streamData.originalPositions, 3));
    geometry.setAttribute('clusterCenter', new THREE.Float32BufferAttribute(streamData.rotationCenters, 3));
    geometry.setAttribute('rotationAxis', new THREE.Float32BufferAttribute(streamData.rotationAxes, 3));
    geometry.setAttribute('rotationSpeed', new THREE.Float32BufferAttribute(streamData.rotationSpeeds, 1));
    geometry.setAttribute('groupCenter', new THREE.Float32BufferAttribute(streamData.groupCenters, 3));
    geometry.setAttribute('orbitalSpeed', new THREE.Float32BufferAttribute(streamData.orbitalSpeeds, 1));

    const material = createGalaxyShaderMaterial(GALAXY_SHADER_PRESETS.galaxyGroup);

    return new THREE.Points(geometry, material);
  }, [streamData]);

  // Animation loop - now only updates shader time uniform
  useFrame((state) => {
    // Update shader time uniform for GPU-based dual rotation and pulsing effects
    if (streamPoints && streamPoints.material instanceof THREE.ShaderMaterial) {
      streamPoints.material.uniforms.time.value = state.clock.elapsedTime;
    }
  });

  return (
    <group ref={groupRef}>
      {/* Galaxy group with individual galaxies and orbital motion */}
      {streamPoints && (
        <primitive object={streamPoints} />
      )}
    </group>
  );
}