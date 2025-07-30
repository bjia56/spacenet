import * as THREE from 'three';

export interface GalaxyShaderOptions {
  /** Time frequency for pulsing animation (default: 2.0) */
  pulseFrequency?: number;
  /** Base pulse amplitude (default: 0.2) */
  pulseAmplitude?: number;
  /** Base pulse offset (default: 0.8) */
  pulseBase?: number;
  /** Position-based frequency for variation (default: 0.1) */
  positionFrequency?: number;
  /** Size scaling factor (default: 300.0) */
  sizeScale?: number;
  /** Inner glow intensity (default: 2.0) */
  innerGlowIntensity?: number;
  /** Outer glow intensity (default: 1.5) */
  outerGlowIntensity?: number;
  /** Alpha falloff power (default: 2.0 for quadratic) */
  alphaFalloff?: number;
}

export function createGalaxyShaderMaterial(options: GalaxyShaderOptions = {}): THREE.ShaderMaterial {
  const {
    pulseFrequency = 2.0,
    pulseAmplitude = 0.2,
    pulseBase = 0.8,
    positionFrequency = 0.1,
    sizeScale = 300.0,
    innerGlowIntensity = 2.0,
    outerGlowIntensity = 1.5,
    alphaFalloff = 2.0
  } = options;

  return new THREE.ShaderMaterial({
    uniforms: {
      time: { value: 0 }
    },
    vertexShader: `
      attribute float size;
      varying vec3 vColor;
      varying float vSize;
      uniform float time;

      void main() {
        vColor = color;
        vSize = size;

        vec4 mvPosition = modelViewMatrix * vec4(position, 1.0);

        // Add subtle animation based on time and position
        float pulse = ${pulseBase.toFixed(1)} + ${pulseAmplitude.toFixed(1)} * sin(time * ${pulseFrequency.toFixed(1)} + position.x * ${positionFrequency.toFixed(1)} + position.y * ${positionFrequency.toFixed(1)} + position.z * ${positionFrequency.toFixed(1)});
        gl_PointSize = size * pulse * (${sizeScale.toFixed(1)} / -mvPosition.z);

        gl_Position = projectionMatrix * mvPosition;
      }
    `,
    fragmentShader: `
      varying vec3 vColor;
      varying float vSize;

      void main() {
        // Create circular point with soft edges
        vec2 center = gl_PointCoord - vec2(0.5);
        float dist = length(center);

        // Soft circular falloff
        float alpha = 1.0 - smoothstep(0.0, 0.5, dist);

        // Add glowing effects
        float innerGlow = 1.0 - smoothstep(0.0, 0.2, dist);
        float outerGlow = 1.0 - smoothstep(0.2, 0.4, dist);
        vec3 glowColor = vColor * (1.0 + innerGlow * ${innerGlowIntensity.toFixed(1)} + outerGlow * ${outerGlowIntensity.toFixed(1)});

        // Apply alpha falloff
        ${alphaFalloff === 2.0 ? 'alpha *= alpha;' : alphaFalloff === 3.0 ? 'alpha *= alpha * alpha;' : `alpha = pow(alpha, ${alphaFalloff.toFixed(1)});`}

        gl_FragColor = vec4(glowColor, alpha);
      }
    `,
    transparent: true,
    blending: THREE.AdditiveBlending,
    depthWrite: false,
    vertexColors: true
  });
}

// Preset configurations for different cosmic structures
export const GALAXY_SHADER_PRESETS = {
  /** For cosmic web filaments - more subtle and flowing */
  cosmicWeb: {
    pulseFrequency: 2.0,
    pulseAmplitude: 0.2,
    pulseBase: 0.8,
    positionFrequency: 0.1,
    sizeScale: 300.0,
    innerGlowIntensity: 2.0,
    outerGlowIntensity: 1.0,
    alphaFalloff: 2.0
  } as GalaxyShaderOptions,

  /** For superclusters - brighter and more dramatic */
  supercluster: {
    pulseFrequency: 1.5,
    pulseAmplitude: 0.3,
    pulseBase: 0.7,
    positionFrequency: 0.05,
    sizeScale: 200.0,
    innerGlowIntensity: 3.0,
    outerGlowIntensity: 1.5,
    alphaFalloff: 3.0
  } as GalaxyShaderOptions,

  /** For galaxy clusters - moderate glow */
  cluster: {
    pulseFrequency: 1.8,
    pulseAmplitude: 0.25,
    pulseBase: 0.75,
    positionFrequency: 0.08,
    sizeScale: 250.0,
    innerGlowIntensity: 2.5,
    outerGlowIntensity: 1.2,
    alphaFalloff: 2.5
  } as GalaxyShaderOptions
};

export function createWavePropagationShaderMaterial(
  numFilaments: number,
  options: GalaxyShaderOptions = {}
): THREE.ShaderMaterial {
  const {
    pulseFrequency = 2.0,
    pulseAmplitude = 0.2,
    pulseBase = 0.8,
    positionFrequency = 0.1,
    sizeScale = 300.0,
    innerGlowIntensity = 2.0,
    outerGlowIntensity = 1.5,
    alphaFalloff = 2.0
  } = options;

  return new THREE.ShaderMaterial({
    uniforms: {
      time: { value: 0 },
      waveSpeed: { value: 0.5 }, // Wave propagation speed
      waveFrequency: { value: 3.0 }, // Wave frequency
      waveAmplitude: { value: 0.8 } // Wave amplitude multiplier
    },
    vertexShader: `
      attribute float size;
      attribute float wavePosition;
      attribute float filamentId;
      varying vec3 vColor;
      varying float vSize;
      varying float vWaveIntensity;
      uniform float time;
      uniform float waveSpeed;
      uniform float waveFrequency;
      uniform float waveAmplitude;

      void main() {
        vColor = color;
        vSize = size;

        // Calculate wave intensity based on position along filament
        float wavePhase = wavePosition * waveFrequency + time * waveSpeed;
        float waveIntensity = sin(wavePhase) * 0.5 + 0.5; // 0 to 1 range
        vWaveIntensity = waveIntensity;

        // Add subtle bobbing motion perpendicular to filament
        // Create pseudo-random direction based on position
        float hash = sin(position.x * 12.9898 + position.y * 78.233 + position.z * 37.719) * 43758.5453;
        float bobFreq = fract(hash) * 0.5; // Random frequency for bobbing

        // Calculate bobbing displacement
        float bobPhase = time * bobFreq + hash;
        float bobAmplitude = 0.15; // Subtle movement
        float bobOffset = sin(bobPhase) * bobAmplitude;

        // Create perpendicular direction for bobbing
        vec3 up = vec3(0.0, 1.0, 0.0);
        vec3 forward = normalize(vec3(sin(wavePosition * 6.28), 0.5, cos(wavePosition * 6.28)));
        vec3 right = normalize(cross(forward, up));
        vec3 perpendicular = normalize(cross(forward, right));

        // Apply bobbing displacement
        vec3 bobbingPosition = position + perpendicular * bobOffset;

        vec4 mvPosition = modelViewMatrix * vec4(bobbingPosition, 1.0);

        // Combine base pulse with wave effect
        float basePulse = ${pulseBase.toFixed(1)} + ${pulseAmplitude.toFixed(1)} * sin(time * ${pulseFrequency.toFixed(1)} + bobbingPosition.x * ${positionFrequency.toFixed(1)} + bobbingPosition.y * ${positionFrequency.toFixed(1)} + bobbingPosition.z * ${positionFrequency.toFixed(1)});
        float wavePulse = 1.0 + waveIntensity * waveAmplitude;

        gl_PointSize = size * basePulse * wavePulse * (${sizeScale.toFixed(1)} / -mvPosition.z);

        gl_Position = projectionMatrix * mvPosition;
      }
    `,
    fragmentShader: `
      varying vec3 vColor;
      varying float vSize;
      varying float vWaveIntensity;

      void main() {
        // Create circular point with soft edges
        vec2 center = gl_PointCoord - vec2(0.5);
        float dist = length(center);

        // Soft circular falloff
        float alpha = 1.0 - smoothstep(0.0, 0.5, dist);

        // Add glowing effects enhanced by wave
        float innerGlow = 1.0 - smoothstep(0.0, 0.2, dist);
        float outerGlow = 1.0 - smoothstep(0.2, 0.4, dist);

        // Wave enhancement - more subtle brightness boost
        float waveBoost = 1.0 + vWaveIntensity * 0.8;
        vec3 glowColor = vColor * waveBoost * (1.0 + innerGlow * ${innerGlowIntensity.toFixed(1)} + outerGlow * ${outerGlowIntensity.toFixed(1)});

        // Apply alpha falloff
        ${alphaFalloff === 2.0 ? 'alpha *= alpha;' : alphaFalloff === 3.0 ? 'alpha *= alpha * alpha;' : `alpha = pow(alpha, ${alphaFalloff.toFixed(1)});`}

        // Wave also affects transparency - more subtle
        alpha *= (0.85 + vWaveIntensity * 0.15);

        gl_FragColor = vec4(glowColor, alpha);
      }
    `,
    transparent: true,
    blending: THREE.AdditiveBlending,
    depthWrite: false,
    vertexColors: true
  });
}