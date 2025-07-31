'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';
import { createGalaxyShaderMaterial, GALAXY_SHADER_PRESETS } from '@/shaders/galaxyShaders';

interface Galaxy3DProps {
  ipSeed: number;
}

export function Galaxy3D({ ipSeed }: Galaxy3DProps) {
  const groupRef = useRef<THREE.Group>(null);

  // Generate galaxy parameters based on IP seed
  const galaxyParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);

    return {
      numArms: 4 + Math.floor(rng.random() * 3), // 4-6 arms
      pointsPerArm: 1000 + Math.floor(rng.random() * 500), // 1000-1500 points per arm
      interArmPoints: 2000 + Math.floor(rng.random() * 1000), // 2000-3000 sparse inter-arm points
      haloPoints: 600 + Math.floor(rng.random() * 400), // 600-1000 halo particles above/below
      // Supermassive black hole parameters
      blackHole: {
        isActive: rng.random() > 0.7, // 30% chance of active galactic nucleus
        accretionDiskSize: 0.3 + rng.random() * 0.4, // 0.3-0.7 size for accretion disk
        jetIntensity: rng.random(), // 0-1 intensity for relativistic jets
        activityLevel: rng.random(), // How bright/active the AGN is
      },
      size: 0.8 + rng.random() * 0.4, // 0.8-1.2 size multiplier
      armTightness: 0.3 + rng.random() * 0.15, // 0.3-0.45 arm tightness
      spinSpeed: 0.02 + rng.random() * 0.01, // 0.02-0.03 rotation speed
      isClockwise: rng.random() > 0.5,
      // Random galaxy tilt - gives each galaxy a unique orientation
      tilt: {
        x: (rng.random() - 0.5) * Math.PI * 0.8, // ±72 degrees max tilt on X axis
        y: (rng.random() - 0.5) * Math.PI * 0.4, // ±36 degrees max tilt on Y axis
        z: (rng.random() - 0.5) * Math.PI * 0.6, // ±54 degrees max tilt on Z axis
      },
      colors: {
        core: new THREE.Color().setHSL(0.1 + rng.random() * 0.05, 0.8, 0.7), // Red-orange core
        stars: new THREE.Color().setHSL(0.15 + rng.random() * 0.05, 0.7, 0.8), // Yellow-white stars
        arms: new THREE.Color().setHSL(0.6 + rng.random() * 0.1, 0.8, 0.6), // Blue arms
        dust: new THREE.Color().setHSL(0.65 + rng.random() * 0.05, 0.5, 0.3), // Dark dust
        interArm: new THREE.Color().setHSL(0.08 + rng.random() * 0.04, 0.4, 0.4), // Dim reddish inter-arm
        halo: new THREE.Color().setHSL(0.05 + rng.random() * 0.03, 0.3, 0.25), // Very dim reddish halo
        // Black hole / AGN colors
        accretionDisk: new THREE.Color().setHSL(0.05 + rng.random() * 0.1, 1.0, 0.9), // Brilliant orange-white
        jets: new THREE.Color().setHSL(0.6 + rng.random() * 0.1, 0.8, 0.8), // Bright blue jets
      }
    };
  }, [ipSeed]);

  // Generate core geometry with GPU rotation attributes
  const coreData = useMemo(() => {
    const rng = new SeededRandom(ipSeed);
    const corePoints = 200;
    const positions = new Float32Array(corePoints * 3);
    const colors = new Float32Array(corePoints * 3);
    const sizes = new Float32Array(corePoints);

    // GPU rotation attributes
    const originalPositions = new Float32Array(corePoints * 3);
    const rotationCenters = new Float32Array(corePoints * 3);
    const rotationAxes = new Float32Array(corePoints * 3);
    const rotationSpeeds = new Float32Array(corePoints);

    const galaxyCenter = new THREE.Vector3(0, 0, 0);
    const rotationAxis = new THREE.Vector3(0, 1, 0); // Y-axis rotation

    for (let i = 0; i < corePoints; i++) {
      // Spherical distribution with higher density toward center
      // But avoid the very center if black hole is present
      const minRadius = galaxyParams.blackHole.isActive ? 0.1 : 0.05; // Keep stars away from black hole
      const radius = minRadius + Math.pow(rng.random(), 2) * (2 * galaxyParams.size - minRadius);
      const theta = rng.random() * Math.PI * 2;
      const phi = Math.acos(2 * rng.random() - 1);

      const x = radius * Math.sin(phi) * Math.cos(theta);
      const y = radius * Math.sin(phi) * Math.sin(theta) * 0.3; // Flatten
      const z = radius * Math.cos(phi) * 0.3;

      positions[i * 3] = x;
      positions[i * 3 + 1] = y;
      positions[i * 3 + 2] = z;

      // Store original positions for GPU rotation
      originalPositions[i * 3] = x;
      originalPositions[i * 3 + 1] = y;
      originalPositions[i * 3 + 2] = z;

      // Set rotation center (galaxy center)
      rotationCenters[i * 3] = galaxyCenter.x;
      rotationCenters[i * 3 + 1] = galaxyCenter.y;
      rotationCenters[i * 3 + 2] = galaxyCenter.z;

      // Set rotation axis
      rotationAxes[i * 3] = rotationAxis.x;
      rotationAxes[i * 3 + 1] = rotationAxis.y;
      rotationAxes[i * 3 + 2] = rotationAxis.z;

      // Black hole gravitational influence: much faster orbits closer to center
      const normalizedRadius = radius / (2 * galaxyParams.size);
      const baseSpeed = galaxyParams.isClockwise ? galaxyParams.spinSpeed : -galaxyParams.spinSpeed;

      // Black hole creates Keplerian orbital motion (faster closer in)
      const blackHoleInfluence = galaxyParams.blackHole ? 3.0 / Math.max(normalizedRadius, 0.1) : 1.0;
      const differentialSpeed = baseSpeed * (1.5 - normalizedRadius * 0.8) * Math.min(blackHoleInfluence, 4.0);
      rotationSpeeds[i] = differentialSpeed;

      // Core colors - brighter toward center
      const distanceFromCenter = radius / (2 * galaxyParams.size);
      const brightness = 1.0 - distanceFromCenter * 0.5;
      const color = galaxyParams.colors.core.clone().multiplyScalar(brightness);

      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;

      // Size based on brightness
      sizes[i] = 1.5 + brightness * 0.8;
    }

    return {
      positions,
      colors,
      sizes,
      originalPositions,
      rotationCenters,
      rotationAxes,
      rotationSpeeds
    };
  }, [ipSeed, galaxyParams]);

  // Generate spiral arm geometries with GPU rotation attributes
  const armData = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 100);
    const armDataSets = [];

    for (let arm = 0; arm < galaxyParams.numArms; arm++) {
      const positions = new Float32Array(galaxyParams.pointsPerArm * 3);
      const colors = new Float32Array(galaxyParams.pointsPerArm * 3);
      const sizes = new Float32Array(galaxyParams.pointsPerArm);

      // GPU rotation attributes
      const originalPositions = new Float32Array(galaxyParams.pointsPerArm * 3);
      const rotationCenters = new Float32Array(galaxyParams.pointsPerArm * 3);
      const rotationAxes = new Float32Array(galaxyParams.pointsPerArm * 3);
      const rotationSpeeds = new Float32Array(galaxyParams.pointsPerArm);

      const galaxyCenter = new THREE.Vector3(0, 0, 0);
      const rotationAxis = new THREE.Vector3(0, 1, 0);

      const armAngle = (arm * 2 * Math.PI) / galaxyParams.numArms;
      const armSeedBase = rng.random();
      const armSeedTwist = rng.random();

      for (let p = 0; p < galaxyParams.pointsPerArm; p++) {
        const progress = p / galaxyParams.pointsPerArm;

        // Spiral radius
        const r = progress * 8 * galaxyParams.size;

        // Spiral angle with wave variations
        const armWave = 0.2 * (0.8 + 0.4 * armSeedBase) *
                       Math.sin(progress * 4) * (1.0 - progress * 0.5);
        const spiralTwist = Math.pow(progress, 0.6 + 0.2 * armSeedTwist);

        const twistDir = galaxyParams.isClockwise ? 1 : -1;
        const theta = armAngle + twistDir * r * galaxyParams.armTightness * spiralTwist + armWave;

        // Add some random scatter
        const scatter = (rng.random() - 0.5) * 0.5;
        const x = r * Math.cos(theta) + scatter;
        const z = r * Math.sin(theta) + scatter;
        const y = (rng.random() - 0.5) * 0.8 * (1.0 - progress); // Disk thickness

        positions[p * 3] = x;
        positions[p * 3 + 1] = y;
        positions[p * 3 + 2] = z;

        // Store original positions for GPU rotation
        originalPositions[p * 3] = x;
        originalPositions[p * 3 + 1] = y;
        originalPositions[p * 3 + 2] = z;

        // Set rotation center (galaxy center)
        rotationCenters[p * 3] = galaxyCenter.x;
        rotationCenters[p * 3 + 1] = galaxyCenter.y;
        rotationCenters[p * 3 + 2] = galaxyCenter.z;

        // Set rotation axis
        rotationAxes[p * 3] = rotationAxis.x;
        rotationAxes[p * 3 + 1] = rotationAxis.y;
        rotationAxes[p * 3 + 2] = rotationAxis.z;

        // Differential rotation: inner parts rotate faster (realistic galactic dynamics)
        const normalizedRadius = r / (8 * galaxyParams.size);
        const baseSpeed = galaxyParams.isClockwise ? galaxyParams.spinSpeed : -galaxyParams.spinSpeed;
        const differentialSpeed = baseSpeed * (2.0 - normalizedRadius * 1.2); // 80%-200% of base speed
        rotationSpeeds[p] = differentialSpeed;

        // Determine star type and color
        let color;
        const starType = rng.random();
        const distanceFromCenter = r / (8 * galaxyParams.size);

        if (p < galaxyParams.pointsPerArm / 6) {
          // Inner region - core stars
          color = galaxyParams.colors.core;
        } else if (starType > 0.8) {
          // Bright stars
          color = galaxyParams.colors.stars;
        } else if (starType > 0.5) {
          // Normal arm stars
          color = galaxyParams.colors.arms;
        } else {
          // Dust and dim stars
          const brightness = 0.5 - distanceFromCenter * 0.3;
          color = galaxyParams.colors.dust.clone().multiplyScalar(brightness);
        }

        colors[p * 3] = color.r;
        colors[p * 3 + 1] = color.g;
        colors[p * 3 + 2] = color.b;

        // Size variation based on star type and distance
        let pointSize;
        if (starType > 0.8) {
          pointSize = 0.4 + rng.random() * 0.4; // Bright stars
        } else if (starType > 0.5) {
          pointSize = 0.2 + rng.random() * 0.4; // Normal stars
        } else {
          pointSize = rng.random() * 0.4; // Dust/dim stars
        }
        sizes[p] = pointSize;
      }

      armDataSets.push({
        positions,
        colors,
        sizes,
        originalPositions,
        rotationCenters,
        rotationAxes,
        rotationSpeeds
      });
    }

    return armDataSets;
  }, [ipSeed, galaxyParams]);

  // Generate inter-arm sparse points to fill space between spiral arms
  const interArmData = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 500);
    const interArmPoints = galaxyParams.interArmPoints;
    const positions = new Float32Array(interArmPoints * 3);
    const colors = new Float32Array(interArmPoints * 3);
    const sizes = new Float32Array(interArmPoints);

    // GPU rotation attributes
    const originalPositions = new Float32Array(interArmPoints * 3);
    const rotationCenters = new Float32Array(interArmPoints * 3);
    const rotationAxes = new Float32Array(interArmPoints * 3);
    const rotationSpeeds = new Float32Array(interArmPoints);

    const galaxyCenter = new THREE.Vector3(0, 0, 0);
    const rotationAxis = new THREE.Vector3(0, 1, 0);

    for (let i = 0; i < interArmPoints; i++) {
      // Random radial distance with bias toward outer regions
      const r = (2 + Math.pow(rng.random(), 0.8) * 6) * galaxyParams.size; // 2-8 units from center

      // Random angle - uniformly distributed
      let theta = rng.random() * Math.PI * 2;

      // Check if this point is in an inter-arm region
      // Calculate expected spiral arm positions at this radius
      const armPositions = [];
      for (let arm = 0; arm < galaxyParams.numArms; arm++) {
        const armAngle = (arm * 2 * Math.PI) / galaxyParams.numArms;
        const twistDir = galaxyParams.isClockwise ? 1 : -1;
        const expectedArmAngle = armAngle + twistDir * r * galaxyParams.armTightness;
        armPositions.push(expectedArmAngle);
      }

      // Find distance to nearest arm
      let minArmDistance = Infinity;
      for (const armAngle of armPositions) {
        const angleDiff = Math.abs(theta - armAngle);
        const wrappedDiff = Math.min(angleDiff, 2 * Math.PI - angleDiff);
        minArmDistance = Math.min(minArmDistance, wrappedDiff);
      }

      // Only place points that are sufficiently far from spiral arms
      const minArmSeparation = 0.4; // Minimum angular distance from arms
      if (minArmDistance < minArmSeparation) {
        // Skip this point or move it to inter-arm region
        const armAvoidanceShift = (minArmSeparation - minArmDistance + 0.2) * (rng.random() > 0.5 ? 1 : -1);
        theta += armAvoidanceShift;
      }

      // Add some random scatter
      const scatter = (rng.random() - 0.5) * 0.8;
      const x = r * Math.cos(theta) + scatter;
      const z = r * Math.sin(theta) + scatter;
      const y = (rng.random() - 0.5) * 1.2 * (1.0 - r / (8 * galaxyParams.size)); // Thicker disk at center

      positions[i * 3] = x;
      positions[i * 3 + 1] = y;
      positions[i * 3 + 2] = z;

      // Store original positions for GPU rotation
      originalPositions[i * 3] = x;
      originalPositions[i * 3 + 1] = y;
      originalPositions[i * 3 + 2] = z;

      // Set rotation center (galaxy center)
      rotationCenters[i * 3] = galaxyCenter.x;
      rotationCenters[i * 3 + 1] = galaxyCenter.y;
      rotationCenters[i * 3 + 2] = galaxyCenter.z;

      // Set rotation axis
      rotationAxes[i * 3] = rotationAxis.x;
      rotationAxes[i * 3 + 1] = rotationAxis.y;
      rotationAxes[i * 3 + 2] = rotationAxis.z;

      // Differential rotation similar to arms but slightly slower
      const normalizedRadius = r / (8 * galaxyParams.size);
      const baseSpeed = galaxyParams.isClockwise ? galaxyParams.spinSpeed : -galaxyParams.spinSpeed;
      const differentialSpeed = baseSpeed * (1.8 - normalizedRadius * 1.0) * 0.9; // 90% of arm speed
      rotationSpeeds[i] = differentialSpeed;

      // Inter-arm colors - dimmer, older stellar populations
      const distanceFromCenter = r / (8 * galaxyParams.size);
      const starType = rng.random();
      let color;

      if (starType > 0.9) {
        // 10% - Occasional bright star
        color = galaxyParams.colors.stars.clone().multiplyScalar(0.6);
      } else if (starType > 0.7) {
        // 20% - Normal inter-arm stars
        color = galaxyParams.colors.interArm.clone().multiplyScalar(0.8 + rng.random() * 0.4);
      } else {
        // 70% - Dim dust and old stars
        const brightness = 0.3 - distanceFromCenter * 0.2;
        color = galaxyParams.colors.dust.clone().multiplyScalar(Math.max(0.1, brightness));
      }

      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;

      // Smaller, dimmer points for inter-arm regions
      const pointSize = starType > 0.9 ?
        0.3 + rng.random() * 0.2 : // Bright stars
        0.1 + rng.random() * 0.2;   // Dim stars
      sizes[i] = pointSize;
    }

    return {
      positions,
      colors,
      sizes,
      originalPositions,
      rotationCenters,
      rotationAxes,
      rotationSpeeds
    };
  }, [ipSeed, galaxyParams]);

  // Generate galactic halo particles above and below the disk
  const haloData = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 800);
    const haloPoints = galaxyParams.haloPoints;
    const positions = new Float32Array(haloPoints * 3);
    const colors = new Float32Array(haloPoints * 3);
    const sizes = new Float32Array(haloPoints);

    // GPU rotation attributes
    const originalPositions = new Float32Array(haloPoints * 3);
    const rotationCenters = new Float32Array(haloPoints * 3);
    const rotationAxes = new Float32Array(haloPoints * 3);
    const rotationSpeeds = new Float32Array(haloPoints);

    const galaxyCenter = new THREE.Vector3(0, 0, 0);
    const rotationAxis = new THREE.Vector3(0, 1, 0);

    for (let i = 0; i < haloPoints; i++) {
      // Create spheroidal halo distribution extending above and below the disk
      const u = rng.random(); // For radius distribution
      const v = rng.random(); // For height distribution
      const w = rng.random(); // For angle

      // Radial distance - bias toward outer regions but include some inner halo
      const r = (3 + Math.pow(u, 0.6) * 12) * galaxyParams.size; // 3-15 units from center

      // Height above/below disk - more concentrated near disk, extending far out
      const maxHeight = 8 * galaxyParams.size; // Maximum halo height
      const heightBias = Math.pow(v, 2); // Bias toward disk plane
      const height = (rng.random() - 0.5) * 2 * maxHeight * heightBias;

      // Random angle around galaxy
      const theta = w * Math.PI * 2;

      // Create slightly flattened spheroidal distribution
      const diskRadius = r * (0.7 + 0.3 * Math.abs(height) / maxHeight); // Smaller radius at greater heights

      const x = diskRadius * Math.cos(theta);
      const z = diskRadius * Math.sin(theta);
      const y = height;

      positions[i * 3] = x;
      positions[i * 3 + 1] = y;
      positions[i * 3 + 2] = z;

      // Store original positions for GPU rotation
      originalPositions[i * 3] = x;
      originalPositions[i * 3 + 1] = y;
      originalPositions[i * 3 + 2] = z;

      // Set rotation center (galaxy center)
      rotationCenters[i * 3] = galaxyCenter.x;
      rotationCenters[i * 3 + 1] = galaxyCenter.y;
      rotationCenters[i * 3 + 2] = galaxyCenter.z;

      // Set rotation axis
      rotationAxes[i * 3] = rotationAxis.x;
      rotationAxes[i * 3 + 1] = rotationAxis.y;
      rotationAxes[i * 3 + 2] = rotationAxis.z;

      // Halo rotation - much slower than disk, more like dark matter halo
      const distanceFromCenter = Math.sqrt(x*x + y*y + z*z);
      const normalizedDistance = distanceFromCenter / (15 * galaxyParams.size);
      const baseSpeed = galaxyParams.isClockwise ? galaxyParams.spinSpeed : -galaxyParams.spinSpeed;
      const haloSpeed = baseSpeed * (0.3 - normalizedDistance * 0.2) * 0.4; // Much slower, 20-50% of disk speed
      rotationSpeeds[i] = haloSpeed;

      // Halo colors - very old, metal-poor stellar populations
      const totalDistance = Math.sqrt(x*x + y*y + z*z);
      const starType = rng.random();
      let color;

      if (starType > 0.95) {
        // 5% - Rare bright halo giants (old red giants, RR Lyrae variables)
        const brightness = 0.4 - (totalDistance / (15 * galaxyParams.size)) * 0.3;
        color = galaxyParams.colors.core.clone().multiplyScalar(Math.max(0.1, brightness));
      } else if (starType > 0.85) {
        // 10% - Metal-poor main sequence stars
        const brightness = 0.3 - (totalDistance / (15 * galaxyParams.size)) * 0.25;
        color = galaxyParams.colors.stars.clone().multiplyScalar(Math.max(0.05, brightness * 0.6));
      } else {
        // 85% - Very dim old halo stars
        const brightness = 0.2 - (totalDistance / (15 * galaxyParams.size)) * 0.15;
        color = galaxyParams.colors.halo.clone().multiplyScalar(Math.max(0.02, brightness));
      }

      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;

      // Very small, dim points for halo
      const pointSize = starType > 0.95 ?
        0.2 + rng.random() * 0.15 : // Rare bright stars
        0.05 + rng.random() * 0.1;   // Most halo stars
      sizes[i] = pointSize;
    }

    return {
      positions,
      colors,
      sizes,
      originalPositions,
      rotationCenters,
      rotationAxes,
      rotationSpeeds
    };
  }, [ipSeed, galaxyParams]);

  // Generate black hole accretion disk and jets (if active)
  const blackHoleData = useMemo(() => {
    if (!galaxyParams.blackHole.isActive) {
      return { positions: new Float32Array(0), colors: new Float32Array(0), sizes: new Float32Array(0) };
    }

    const rng = new SeededRandom(ipSeed + 1000);
    const diskPoints = 150; // Accretion disk particles
    const jetPoints = galaxyParams.blackHole.jetIntensity > 0.5 ? 100 : 0; // Jets only for high activity
    const totalPoints = diskPoints + jetPoints;

    const positions = new Float32Array(totalPoints * 3);
    const colors = new Float32Array(totalPoints * 3);
    const sizes = new Float32Array(totalPoints);

    // GPU rotation attributes
    const originalPositions = new Float32Array(totalPoints * 3);
    const rotationCenters = new Float32Array(totalPoints * 3);
    const rotationAxes = new Float32Array(totalPoints * 3);
    const rotationSpeeds = new Float32Array(totalPoints);

    const galaxyCenter = new THREE.Vector3(0, 0, 0);
    const rotationAxis = new THREE.Vector3(0, 1, 0);
    let pointIndex = 0;

    // Generate accretion disk
    for (let i = 0; i < diskPoints; i++) {
      // Thin disk around black hole
      const diskRadius = galaxyParams.blackHole.accretionDiskSize;
      const r = (0.02 + Math.pow(rng.random(), 0.3) * diskRadius) * galaxyParams.size;
      const theta = rng.random() * Math.PI * 2;

      // Very thin disk
      const x = r * Math.cos(theta);
      const z = r * Math.sin(theta);
      const y = (rng.random() - 0.5) * 0.02 * galaxyParams.size; // Extremely thin

      positions[pointIndex * 3] = x;
      positions[pointIndex * 3 + 1] = y;
      positions[pointIndex * 3 + 2] = z;

      originalPositions[pointIndex * 3] = x;
      originalPositions[pointIndex * 3 + 1] = y;
      originalPositions[pointIndex * 3 + 2] = z;

      rotationCenters[pointIndex * 3] = galaxyCenter.x;
      rotationCenters[pointIndex * 3 + 1] = galaxyCenter.y;
      rotationCenters[pointIndex * 3 + 2] = galaxyCenter.z;

      rotationAxes[pointIndex * 3] = rotationAxis.x;
      rotationAxes[pointIndex * 3 + 1] = rotationAxis.y;
      rotationAxes[pointIndex * 3 + 2] = rotationAxis.z;

      // Extremely fast rotation near black hole (relativistic speeds)
      const normalizedRadius = r / (diskRadius * galaxyParams.size);
      const baseSpeed = galaxyParams.isClockwise ? galaxyParams.spinSpeed : -galaxyParams.spinSpeed;
      const relativisticSpeed = baseSpeed * (8.0 / Math.max(normalizedRadius, 0.05)); // Extremely fast
      rotationSpeeds[pointIndex] = Math.min(relativisticSpeed, baseSpeed * 20); // Cap at 20x base speed

      // Accretion disk colors - extremely hot and bright
      const temperature = 1.0 - normalizedRadius * 0.7; // Hotter closer to black hole
      const brightness = galaxyParams.blackHole.activityLevel * (0.8 + temperature * 0.4);
      const color = galaxyParams.colors.accretionDisk.clone().multiplyScalar(brightness);

      colors[pointIndex * 3] = color.r;
      colors[pointIndex * 3 + 1] = color.g;
      colors[pointIndex * 3 + 2] = color.b;

      // Bright, variable sizes for accretion disk
      sizes[pointIndex] = (0.5 + temperature * 1.0) * (0.8 + rng.random() * 0.4);
      pointIndex++;
    }

    // Generate relativistic jets (if present)
    for (let i = 0; i < jetPoints; i++) {
      // Jets extend perpendicular to accretion disk
      const jetLength = 5 * galaxyParams.size * galaxyParams.blackHole.jetIntensity;
      const jetRadius = 0.1 * galaxyParams.size;

      // Two jets: one up, one down
      const isUpJet = i < jetPoints / 2;
      const jetDirection = isUpJet ? 1 : -1;

      const t = (i % (jetPoints / 2)) / (jetPoints / 2); // 0 to 1 along jet
      const jetDistance = t * jetLength;

      // Slight cone expansion
      const coneRadius = jetRadius * (1 + t * 0.5);
      const localTheta = rng.random() * Math.PI * 2;
      const localR = rng.random() * coneRadius;

      const x = localR * Math.cos(localTheta);
      const z = localR * Math.sin(localTheta);
      const y = jetDirection * jetDistance;

      positions[pointIndex * 3] = x;
      positions[pointIndex * 3 + 1] = y;
      positions[pointIndex * 3 + 2] = z;

      originalPositions[pointIndex * 3] = x;
      originalPositions[pointIndex * 3 + 1] = y;
      originalPositions[pointIndex * 3 + 2] = z;

      rotationCenters[pointIndex * 3] = galaxyCenter.x;
      rotationCenters[pointIndex * 3 + 1] = galaxyCenter.y;
      rotationCenters[pointIndex * 3 + 2] = galaxyCenter.z;

      rotationAxes[pointIndex * 3] = rotationAxis.x;
      rotationAxes[pointIndex * 3 + 1] = rotationAxis.y;
      rotationAxes[pointIndex * 3 + 2] = rotationAxis.z;

      // Jets don't rotate much with galaxy
      rotationSpeeds[pointIndex] = galaxyParams.spinSpeed * 0.1;

      // Jet colors - bright blue, fading with distance
      const jetBrightness = galaxyParams.blackHole.activityLevel * (1.0 - t * 0.8);
      const color = galaxyParams.colors.jets.clone().multiplyScalar(jetBrightness);

      colors[pointIndex * 3] = color.r;
      colors[pointIndex * 3 + 1] = color.g;
      colors[pointIndex * 3 + 2] = color.b;

      // Jet particle sizes
      sizes[pointIndex] = (0.3 + jetBrightness * 0.5) * (0.7 + rng.random() * 0.6);
      pointIndex++;
    }

    return {
      positions,
      colors,
      sizes,
      originalPositions,
      rotationCenters,
      rotationAxes,
      rotationSpeeds
    };
  }, [ipSeed, galaxyParams]);

  // Create Points with GPU shaders
  const corePoints = useMemo(() => {
    if (!coreData.positions.length) return null;

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(coreData.positions, 3));
    geometry.setAttribute('color', new THREE.Float32BufferAttribute(coreData.colors, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(coreData.sizes, 1));

    // Add rotation attributes for GPU-based differential rotation
    geometry.setAttribute('originalPosition', new THREE.Float32BufferAttribute(coreData.originalPositions, 3));
    geometry.setAttribute('clusterCenter', new THREE.Float32BufferAttribute(coreData.rotationCenters, 3));
    geometry.setAttribute('rotationAxis', new THREE.Float32BufferAttribute(coreData.rotationAxes, 3));
    geometry.setAttribute('rotationSpeed', new THREE.Float32BufferAttribute(coreData.rotationSpeeds, 1));

    const material = createGalaxyShaderMaterial(GALAXY_SHADER_PRESETS.spiralGalaxy);

    return new THREE.Points(geometry, material);
  }, [coreData]);

  const armPoints = useMemo(() => {
    return armData.map((armDataSet) => {
      const geometry = new THREE.BufferGeometry();
      geometry.setAttribute('position', new THREE.Float32BufferAttribute(armDataSet.positions, 3));
      geometry.setAttribute('color', new THREE.Float32BufferAttribute(armDataSet.colors, 3));
      geometry.setAttribute('size', new THREE.Float32BufferAttribute(armDataSet.sizes, 1));

      // Add rotation attributes for GPU-based differential rotation
      geometry.setAttribute('originalPosition', new THREE.Float32BufferAttribute(armDataSet.originalPositions, 3));
      geometry.setAttribute('clusterCenter', new THREE.Float32BufferAttribute(armDataSet.rotationCenters, 3));
      geometry.setAttribute('rotationAxis', new THREE.Float32BufferAttribute(armDataSet.rotationAxes, 3));
      geometry.setAttribute('rotationSpeed', new THREE.Float32BufferAttribute(armDataSet.rotationSpeeds, 1));

      const material = createGalaxyShaderMaterial(GALAXY_SHADER_PRESETS.spiralGalaxy);

      return new THREE.Points(geometry, material);
    });
  }, [armData]);

  const interArmPoints = useMemo(() => {
    if (!interArmData.positions.length) return null;

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(interArmData.positions, 3));
    geometry.setAttribute('color', new THREE.Float32BufferAttribute(interArmData.colors, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(interArmData.sizes, 1));

    // Add rotation attributes for GPU-based differential rotation
    geometry.setAttribute('originalPosition', new THREE.Float32BufferAttribute(interArmData.originalPositions, 3));
    geometry.setAttribute('clusterCenter', new THREE.Float32BufferAttribute(interArmData.rotationCenters, 3));
    geometry.setAttribute('rotationAxis', new THREE.Float32BufferAttribute(interArmData.rotationAxes, 3));
    geometry.setAttribute('rotationSpeed', new THREE.Float32BufferAttribute(interArmData.rotationSpeeds, 1));

    const material = createGalaxyShaderMaterial(GALAXY_SHADER_PRESETS.spiralGalaxy);

    return new THREE.Points(geometry, material);
  }, [interArmData]);

  const haloPoints = useMemo(() => {
    if (!haloData.positions.length) return null;

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(haloData.positions, 3));
    geometry.setAttribute('color', new THREE.Float32BufferAttribute(haloData.colors, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(haloData.sizes, 1));

    // Add rotation attributes for GPU-based differential rotation
    geometry.setAttribute('originalPosition', new THREE.Float32BufferAttribute(haloData.originalPositions, 3));
    geometry.setAttribute('clusterCenter', new THREE.Float32BufferAttribute(haloData.rotationCenters, 3));
    geometry.setAttribute('rotationAxis', new THREE.Float32BufferAttribute(haloData.rotationAxes, 3));
    geometry.setAttribute('rotationSpeed', new THREE.Float32BufferAttribute(haloData.rotationSpeeds, 1));

    const material = createGalaxyShaderMaterial(GALAXY_SHADER_PRESETS.spiralGalaxy);

    return new THREE.Points(geometry, material);
  }, [haloData]);

  const blackHolePoints = useMemo(() => {
    if (!blackHoleData.positions.length) return null;

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(blackHoleData.positions, 3));
    geometry.setAttribute('color', new THREE.Float32BufferAttribute(blackHoleData.colors, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(blackHoleData.sizes, 1));

    // Add rotation attributes for GPU-based rotation
    geometry.setAttribute('originalPosition', new THREE.Float32BufferAttribute(blackHoleData.originalPositions, 3));
    geometry.setAttribute('clusterCenter', new THREE.Float32BufferAttribute(blackHoleData.rotationCenters, 3));
    geometry.setAttribute('rotationAxis', new THREE.Float32BufferAttribute(blackHoleData.rotationAxes, 3));
    geometry.setAttribute('rotationSpeed', new THREE.Float32BufferAttribute(blackHoleData.rotationSpeeds, 1));

    const material = createGalaxyShaderMaterial(GALAXY_SHADER_PRESETS.spiralGalaxy);

    return new THREE.Points(geometry, material);
  }, [blackHoleData]);

  // Animation loop - now only updates shader time uniform for GPU rotation
  useFrame((state) => {
    const time = state.clock.elapsedTime;

    // Update shader time uniform for GPU-based differential rotation
    if (corePoints && corePoints.material instanceof THREE.ShaderMaterial) {
      corePoints.material.uniforms.time.value = time;
    }

    // Update spiral arm shader time uniforms
    armPoints.forEach(armPoint => {
      if (armPoint && armPoint.material instanceof THREE.ShaderMaterial) {
        armPoint.material.uniforms.time.value = time;
      }
    });

    // Update inter-arm shader time uniform
    if (interArmPoints && interArmPoints.material instanceof THREE.ShaderMaterial) {
      interArmPoints.material.uniforms.time.value = time;
    }

    // Update halo shader time uniform
    if (haloPoints && haloPoints.material instanceof THREE.ShaderMaterial) {
      haloPoints.material.uniforms.time.value = time;
    }

    // Update black hole accretion disk and jets shader time uniform
    if (blackHolePoints && blackHolePoints.material instanceof THREE.ShaderMaterial) {
      blackHolePoints.material.uniforms.time.value = time;
    }

    // Apply galaxy tilt and add gentle precession (whole galaxy wobble)
    if (groupRef.current) {
      // Apply base tilt from galaxy parameters
      groupRef.current.rotation.x = galaxyParams.tilt.x + Math.sin(time * 0.1) * 0.1;
      groupRef.current.rotation.y = galaxyParams.tilt.y + Math.cos(time * 0.08) * 0.03;
      groupRef.current.rotation.z = galaxyParams.tilt.z + Math.cos(time * 0.15) * 0.05;
    }
  });

  return (
    <group ref={groupRef}>
      {/* Galaxy core with GPU differential rotation */}
      {corePoints && (
        <primitive object={corePoints} />
      )}

      {/* Spiral arms with GPU differential rotation */}
      {armPoints.map((armPoint, index) => (
        <primitive key={index} object={armPoint} />
      ))}

      {/* Inter-arm sparse points */}
      {interArmPoints && (
        <primitive object={interArmPoints} />
      )}

      {/* Galactic halo particles above and below disk */}
      {haloPoints && (
        <primitive object={haloPoints} />
      )}

      {/* Supermassive black hole accretion disk and jets (if active) */}
      {blackHolePoints && (
        <primitive object={blackHolePoints} />
      )}
    </group>
  );
}