// Portions sourced from https://github.com/dgreenheck/threejs-procedural-planets

'use client';

import { useRef, useMemo, useEffect } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { SeededRandom } from '@/lib/seededRandom';

const noiseFunctions = `
const float PI = 3.14159265;

//	Simplex 3D Noise
//	by Ian McEwan, Ashima Arts
//
vec4 permute(vec4 x){return mod(((x*34.0)+1.0)*x, 289.0);}
vec4 taylorInvSqrt(vec4 r){return 1.79284291400159 - 0.85373472095314 * r;}

//
float simplex3(vec3 v) {
  const vec2  C = vec2(1.0/6.0, 1.0/3.0) ;
  const vec4  D = vec4(0.0, 0.5, 1.0, 2.0);

  // First corner
  vec3 i  = floor(v + dot(v, C.yyy) );
  vec3 x0 =   v - i + dot(i, C.xxx) ;

  // Other corners
  vec3 g = step(x0.yzx, x0.xyz);
  vec3 l = 1.0 - g;
  vec3 i1 = min( g.xyz, l.zxy );
  vec3 i2 = max( g.xyz, l.zxy );

  //  x0 = x0 - 0. + 0.0 * C
  vec3 x1 = x0 - i1 + 1.0 * C.xxx;
  vec3 x2 = x0 - i2 + 2.0 * C.xxx;
  vec3 x3 = x0 - 1. + 3.0 * C.xxx;

  // Permutations
  i = mod(i, 289.0 );
  vec4 p = permute( permute( permute(
            i.z + vec4(0.0, i1.z, i2.z, 1.0 ))
          + i.y + vec4(0.0, i1.y, i2.y, 1.0 ))
          + i.x + vec4(0.0, i1.x, i2.x, 1.0 ));

  // Gradients
  // ( N*N points uniformly over a square, mapped onto an octahedron.)
  float n_ = 1.0/7.0; // N=7
  vec3  ns = n_ * D.wyz - D.xzx;

  vec4 j = p - 49.0 * floor(p * ns.z *ns.z);  //  mod(p,N*N)

  vec4 x_ = floor(j * ns.z);
  vec4 y_ = floor(j - 7.0 * x_ );    // mod(j,N)

  vec4 x = x_ *ns.x + ns.yyyy;
  vec4 y = y_ *ns.x + ns.yyyy;
  vec4 h = 1.0 - abs(x) - abs(y);

  vec4 b0 = vec4( x.xy, y.xy );
  vec4 b1 = vec4( x.zw, y.zw );

  vec4 s0 = floor(b0)*2.0 + 1.0;
  vec4 s1 = floor(b1)*2.0 + 1.0;
  vec4 sh = -step(h, vec4(0.0));

  vec4 a0 = b0.xzyw + s0.xzyw*sh.xxyy ;
  vec4 a1 = b1.xzyw + s1.xzyw*sh.zzww ;

  vec3 p0 = vec3(a0.xy,h.x);
  vec3 p1 = vec3(a0.zw,h.y);
  vec3 p2 = vec3(a1.xy,h.z);
  vec3 p3 = vec3(a1.zw,h.w);

  //Normalise gradients
  vec4 norm = taylorInvSqrt(vec4(dot(p0,p0), dot(p1,p1), dot(p2, p2), dot(p3,p3)));
  p0 *= norm.x;
  p1 *= norm.y;
  p2 *= norm.z;
  p3 *= norm.w;

  // Mix final noise value
  vec4 m = max(0.6 - vec4(dot(x0,x0), dot(x1,x1), dot(x2,x2), dot(x3,x3)), 0.0);
  m = m * m;
  return 42.0 * dot( m*m, vec4( dot(p0,x0), dot(p1,x1),
                                dot(p2,x2), dot(p3,x3) ) );
}

float fractal3(
  vec3 v,
  float sharpness,
  float period,
  float persistence,
  float lacunarity,
  int octaves
) {
  float n = 0.0;
  float a = 1.0; // Amplitude for current octave
  float max_amp = 0.0; // Accumulate max amplitude so we can normalize after
  float P = period;  // Period for current octave

  for(int i = 0; i < octaves; i++) {
      n += a * simplex3(v / P);
      a *= persistence;
      max_amp += a;
      P /= lacunarity;
  }

  // Normalize noise between [0.0, amplitude]
  return n / max_amp;
}

float terrainHeight(
  int type,
  vec3 v,
  float amplitude,
  float sharpness,
  float offset,
  float period,
  float persistence,
  float lacunarity,
  int octaves
) {
  float h = 0.0;

  if (type == 1) {
    h = amplitude * simplex3(v / period);
  } else if (type == 2) {
    h = amplitude * fractal3(
      v,
      sharpness,
      period,
      persistence,
      lacunarity,
      octaves);
    h = amplitude * pow(max(0.0, (h + 1.0) / 2.0), sharpness);
  } else if (type == 3) {
    h = fractal3(
      v,
      sharpness,
      period,
      persistence,
      lacunarity,
      octaves);
    h = amplitude * pow(max(0.0, 1.0 - abs(h)), sharpness);
  }

  // Multiply by amplitude and adjust offset
  return max(0.0, h + offset);
}
`;

// Vertex shader for the planet
const planetVertexShader = `
attribute vec3 tangent;

// Terrain generation parameters
uniform int type;
uniform float radius;
uniform float amplitude;
uniform float sharpness;
uniform float offset;
uniform float period;
uniform float persistence;
uniform float lacunarity;
uniform int octaves;

// Bump mapping
uniform float bumpStrength;
uniform float bumpOffset;

varying vec3 fragPosition;
varying vec3 fragNormal;
varying vec3 fragTangent;
varying vec3 fragBitangent;

${noiseFunctions}

void main() {
  // Calculate terrain height
  float h = terrainHeight(
    type,
    position,
    amplitude,
    sharpness,
    offset,
    period,
    persistence,
    lacunarity,
    octaves);

  vec3 pos = position * (radius + h);

  gl_Position = projectionMatrix * modelViewMatrix * vec4(pos, 1.0);
  fragPosition = position;
  fragNormal = normal;
  fragTangent = tangent;
  fragBitangent = cross(normal, tangent);
}
`;

// Fragment shader for the planet
const planetFragmentShader = `
// Terrain generation parameters
uniform int type;
uniform float radius;
uniform float amplitude;
uniform float sharpness;
uniform float offset;
uniform float period;
uniform float persistence;
uniform float lacunarity;
uniform int octaves;

// Layer colors
uniform vec3 color1;
uniform vec3 color2;
uniform vec3 color3;
uniform vec3 color4;
uniform vec3 color5;

// Transition points for each layer
uniform float transition2;
uniform float transition3;
uniform float transition4;
uniform float transition5;

// Amount of blending between each layer
uniform float blend12;
uniform float blend23;
uniform float blend34;
uniform float blend45;

// Bump mapping parameters
uniform float bumpStrength;
uniform float bumpOffset;

// Lighting parameters
uniform float ambientIntensity;
uniform float diffuseIntensity;
uniform float specularIntensity;
uniform float shininess;
uniform vec3 lightPosition;
uniform vec3 lightDirection;
uniform vec3 lightColor;

varying vec3 fragPosition;
varying vec3 fragNormal;
varying vec3 fragTangent;
varying vec3 fragBitangent;

${noiseFunctions}

void main() {
  // Calculate terrain height
  float h = terrainHeight(
    type,
    fragPosition,
    amplitude,
    sharpness,
    offset,
    period,
    persistence,
    lacunarity,
    octaves);

  vec3 dx = bumpOffset * fragTangent;
  float h_dx = terrainHeight(
    type,
    fragPosition + dx,
    amplitude,
    sharpness,
    offset,
    period,
    persistence,
    lacunarity,
    octaves);

  vec3 dy = bumpOffset * fragBitangent;
  float h_dy = terrainHeight(
    type,
    fragPosition + dy,
    amplitude,
    sharpness,
    offset,
    period,
    persistence,
    lacunarity,
    octaves);

  vec3 pos = fragPosition * (radius + h);
  vec3 pos_dx = (fragPosition + dx) * (radius + h_dx);
  vec3 pos_dy = (fragPosition + dy) * (radius + h_dy);

  // Recalculate surface normal post-bump mapping
  vec3 bumpNormal = normalize(cross(pos_dx - pos, pos_dy - pos));
  // Mix original normal and bumped normal to control bump strength
  vec3 N = normalize(mix(fragNormal, bumpNormal, bumpStrength));

  // Normalized light direction (points in direction that light travels)
  vec3 L = normalize(-lightDirection);
  // View vector from light to fragment
  vec3 V = normalize(lightPosition - pos);
  // Reflected light vector
  vec3 R = normalize(reflect(L, N));

  float diffuse = diffuseIntensity * max(0.0, dot(N, -L));

  // https://ogldev.org/www/tutorial19/tutorial19.html
  float specularFalloff = clamp((transition3 - h) / transition3, 0.0, 1.0);
  float specular = max(0.0, specularFalloff * specularIntensity * pow(dot(V, R), shininess));

  float light = ambientIntensity + diffuse + specular;

  // Blender colors layer by layer
  vec3 color12 = mix(
    color1,
    color2,
    smoothstep(transition2 - blend12, transition2 + blend12, h));

  vec3 color123 = mix(
    color12,
    color3,
    smoothstep(transition3 - blend23, transition3 + blend23, h));

  vec3 color1234 = mix(
    color123,
    color4,
    smoothstep(transition4 - blend34, transition4 + blend34, h));

  vec3 finalColor = mix(
    color1234,
    color5,
    smoothstep(transition5 - blend45, transition5 + blend45, h));

  gl_FragColor = vec4(light * finalColor * lightColor, 1.0);
}
`;

// Vertex shader for the atmosphere
const atmosphereVertexShader = `
attribute float size;

varying vec3 fragPosition;

void main() {
  gl_PointSize = size;
  gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
  fragPosition = (modelMatrix * vec4(position, 1.0)).xyz;
}
`;

// Fragment shader for the atmosphere
const atmosphereFragmentShader = `
uniform float time;
uniform float speed;
uniform float opacity;
uniform float density;
uniform float scale;

uniform vec3 lightDirection;

uniform vec3 color;
uniform sampler2D pointTexture;

varying vec3 fragPosition;

vec2 rotateUV(vec2 uv, float rotation) {
    float mid = 0.5;
    return vec2(
        cos(rotation) * (uv.x - mid) + sin(rotation) * (uv.y - mid) + mid,
        cos(rotation) * (uv.y - mid) - sin(rotation) * (uv.x - mid) + mid
    );
}

${noiseFunctions}

void main() {
  vec3 R = normalize(fragPosition);
  vec3 L = normalize(lightDirection);
  float light = max(0.05, dot(R, L));

  float n = simplex3((time * speed) + fragPosition / scale);
  float alpha = opacity * clamp(n + density, 0.0, 1.0);

  vec2 rotCoords = rotateUV(gl_PointCoord, n);
  gl_FragColor = vec4(light * color, alpha) * texture2D(pointTexture, gl_PointCoord);
}
`;

interface Planet3DProps {
  ipSeed: number;
}

export function Planet3D({ ipSeed }: Planet3DProps) {
  const groupRef = useRef<THREE.Group>(null);
  const planetRef = useRef<THREE.Mesh>(null);
  const atmosphereRef = useRef<THREE.Points>(null);
  const cloudTexture = useRef<THREE.Texture | null>(null);
  const clock = useRef(new THREE.Clock());

  // Generate parameters based on seed
  const planetParams = useMemo(() => {
    const rng = new SeededRandom(ipSeed);

    // Generate terrain type based on seed
    const terrainTypeRoll = rng.random();
    let terrainType = 2; // Default to fractal
    if (terrainTypeRoll < 0.50) {
      terrainType = 3; // ridged fractal
    }

    // Generate planet color palette with realistic terrain colors
    const planetType = rng.random();

    let color1, color2, color3, color4, color5, atmosphereColor;

    if (planetType < 0.3) {
      // Ocean world - blues and teals
      const baseHue = 0.5 + rng.random() * 0.15; // Blue-cyan range
      color1 = new THREE.Color().setHSL(baseHue, 0.8, 0.15); // Deep ocean
      color2 = new THREE.Color().setHSL(baseHue + 0.05, 0.7, 0.25); // Shallow water
      color3 = new THREE.Color().setHSL(0.1 + rng.random() * 0.1, 0.6, 0.35); // Sandy shores
      color4 = new THREE.Color().setHSL(0.3, 0.5, 0.4); // Vegetation
      color5 = new THREE.Color().setHSL(0.0, 0.0, 0.85); // Ice caps
      atmosphereColor = new THREE.Color().setHSL(baseHue, 0.4, 0.7);
    } else if (planetType < 0.6) {
      // Desert world - browns and oranges
      const baseHue = 0.05 + rng.random() * 0.08; // Orange-brown range
      color1 = new THREE.Color().setHSL(baseHue, 0.4, 0.2); // Dark soil
      color2 = new THREE.Color().setHSL(baseHue + 0.02, 0.5, 0.3); // Rich earth
      color3 = new THREE.Color().setHSL(baseHue + 0.03, 0.6, 0.45); // Sandy terrain
      color4 = new THREE.Color().setHSL(baseHue + 0.05, 0.7, 0.6); // Light sand
      color5 = new THREE.Color().setHSL(0.0, 0.0, 0.8); // Rocky peaks
      atmosphereColor = new THREE.Color().setHSL(baseHue + 0.1, 0.3, 0.65);
    } else if (planetType < 0.8) {
      // Forest world - greens and browns
      const greenHue = 0.25 + rng.random() * 0.15; // Green range
      const brownHue = 0.08 + rng.random() * 0.05; // Brown range
      color1 = new THREE.Color().setHSL(brownHue, 0.6, 0.2); // Dark soil
      color2 = new THREE.Color().setHSL(brownHue, 0.5, 0.3); // Rich earth
      color3 = new THREE.Color().setHSL(greenHue, 0.6, 0.35); // Forest floor
      color4 = new THREE.Color().setHSL(greenHue - 0.05, 0.7, 0.4); // Dense vegetation
      color5 = new THREE.Color().setHSL(0.0, 0.0, 0.75); // Mountain stone
      atmosphereColor = new THREE.Color().setHSL(greenHue, 0.3, 0.65);
    } else {
      // Volcanic/exotic world - reds and purples
      const baseHue = 0.85 + rng.random() * 0.15; // Red-purple range
      color1 = new THREE.Color().setHSL(baseHue, 0.7, 0.15); // Dark volcanic rock
      color2 = new THREE.Color().setHSL(baseHue + 0.05, 0.6, 0.25); // Mineral deposits
      color3 = new THREE.Color().setHSL(baseHue + 0.1, 0.5, 0.4); // Oxidized terrain
      color4 = new THREE.Color().setHSL(baseHue - 0.1, 0.8, 0.5); // Exotic minerals
      color5 = new THREE.Color().setHSL(0.0, 0.0, 0.9); // Metallic peaks
      atmosphereColor = new THREE.Color().setHSL(baseHue, 0.4, 0.6);
    }

    const tParams = {
      // Terrain parameters
      type: terrainType,
      radius: 1,
      amplitude: ((terrainType == 2 ? 0.25 : 0.05) + rng.random() * 0.05),
      sharpness: 2.6,
      offset: -0.01,
      period: 0.5 + rng.random() * 0.6,
      persistence: 0.5 + rng.random() * 0.05,
      lacunarity: 1.6 + rng.random() * 0.2,
      octaves: 10,

      // Layer colors
      color1,
      color2,
      color3,
      color4,
      color5,

      // Transition points
      transition2: 0.0025,
      transition3: 0.0075,
      transition4: 0.015,
      transition5: 0.05,

      // Blend factors
      blend12: 0.05,
      blend23: 0.05,
      blend34: 0.05,
      blend45: 0.05,

      // Bump mapping
      bumpStrength: 0.05,
      bumpOffset: 0.00005,

      // Lighting
      ambientIntensity: 0.1,
      diffuseIntensity: 1,
      specularIntensity: 2,
      shininess: 10,
      lightPosition: new THREE.Vector3(10, 10, 10),
      lightDirection: new THREE.Vector3(1, 1, 1),
      lightColor: new THREE.Color(0xffffff),

      // Atmosphere
      atmosphereRadius: 1.05,
      atmosphereThickness: 0.02,
      atmosphereParticles: 4000 + Math.floor(rng.random() * 4000),
      minParticleSize: 50,
      maxParticleSize: 100,
      density: 0,
      opacity: 0.35,
      scale: 1,
      speed: 0.01 + rng.random() * 0.05,
      atmosphereColor,

      // Rotation speed
      rotationSpeed: 0.1 + rng.random() * 0.3
    };
    console.log(tParams);
    return tParams;
  }, [ipSeed]);

  // Create a cloud texture
  useEffect(() => {
    const loader = new THREE.TextureLoader();
    const texture = loader.load('https://raw.githubusercontent.com/dgreenheck/threejs-procedural-planets/23457d4f8e6bedf22aa266e12c92f850822ff3a4/public/cloud.png');
    cloudTexture.current = texture;
  }, [ipSeed]);

  // Create atmosphere particles
  const atmosphereGeometry = useMemo(() => {
    const rng = new SeededRandom(ipSeed + 100);
    const geometry = new THREE.BufferGeometry();

    const particles = planetParams.atmosphereParticles;
    const positions = new Float32Array(particles * 3);
    const sizes = new Float32Array(particles);

    for (let i = 0; i < particles; i++) {
      // Generate random point in unit sphere
      const p = new THREE.Vector3(
        2 * rng.random() - 1,
        2 * rng.random() - 1,
        2 * rng.random() - 1
      );

      // Normalize to surface of sphere
      p.normalize();

      // Add random radius within atmosphere thickness
      const radius = planetParams.radius * planetParams.atmosphereRadius +
        rng.random() * planetParams.atmosphereThickness;

      p.multiplyScalar(radius);

      positions[i * 3] = p.x;
      positions[i * 3 + 1] = p.y;
      positions[i * 3 + 2] = p.z;

      // Random particle size
      const minSize = planetParams.minParticleSize;
      const maxSize = planetParams.maxParticleSize;
      sizes[i] = minSize + rng.random() * (maxSize - minSize);
    }

    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('size', new THREE.BufferAttribute(sizes, 1));

    return geometry;
  }, [ipSeed, planetParams]);

  // Animation with key to force remount when seed changes
  const animationKey = useMemo(() => `planet-${ipSeed}`, [ipSeed]);

  useFrame((state, delta) => {
    if (groupRef.current) {
      groupRef.current.rotation.y += delta * planetParams.rotationSpeed * 0.2;
    }

    // Update atmosphere shader time
    if (atmosphereRef.current && atmosphereRef.current.material) {
      (atmosphereRef.current.material as THREE.ShaderMaterial).uniforms.time.value =
        clock.current.getElapsedTime();
    }
  });

  const planetGeometry = useMemo(() => {
    const sphere = new THREE.SphereGeometry(1, 128, 128);
    sphere.computeTangents();
    return sphere;
  }, [ipSeed]);

  return (
    <group ref={groupRef} key={animationKey}>
      {/* Planet with procedural shader */}
      <mesh ref={planetRef}>
        <primitive object={planetGeometry} />
        <shaderMaterial
          key={`planet-material-${ipSeed}`}
          vertexShader={planetVertexShader}
          fragmentShader={planetFragmentShader}
          uniforms={{
            // Terrain parameters
            type: { value: planetParams.type },
            radius: { value: planetParams.radius },
            amplitude: { value: planetParams.amplitude },
            sharpness: { value: planetParams.sharpness },
            offset: { value: planetParams.offset },
            period: { value: planetParams.period },
            persistence: { value: planetParams.persistence },
            lacunarity: { value: planetParams.lacunarity },
            octaves: { value: planetParams.octaves },

            // Layer colors
            color1: { value: planetParams.color1 },
            color2: { value: planetParams.color2 },
            color3: { value: planetParams.color3 },
            color4: { value: planetParams.color4 },
            color5: { value: planetParams.color5 },

            // Transition heights
            transition2: { value: planetParams.transition2 },
            transition3: { value: planetParams.transition3 },
            transition4: { value: planetParams.transition4 },
            transition5: { value: planetParams.transition5 },

            // Blend amounts
            blend12: { value: planetParams.blend12 },
            blend23: { value: planetParams.blend23 },
            blend34: { value: planetParams.blend34 },
            blend45: { value: planetParams.blend45 },

            // Bump mapping
            bumpStrength: { value: planetParams.bumpStrength },
            bumpOffset: { value: planetParams.bumpOffset },

            // Lighting
            ambientIntensity: { value: planetParams.ambientIntensity },
            diffuseIntensity: { value: planetParams.diffuseIntensity },
            specularIntensity: { value: planetParams.specularIntensity },
            shininess: { value: planetParams.shininess },
            lightPosition: { value: planetParams.lightPosition },
            lightDirection: { value: planetParams.lightDirection },
            lightColor: { value: planetParams.lightColor }
          }}
        />
      </mesh>

      {/* Atmospheric particles */}
      <points ref={atmosphereRef} geometry={atmosphereGeometry}>
        <shaderMaterial
          key={`atmosphere-material-${ipSeed}`}
          vertexShader={atmosphereVertexShader}
          fragmentShader={atmosphereFragmentShader}
          uniforms={{
            time: { value: 0 },
            speed: { value: planetParams.speed },
            opacity: { value: planetParams.opacity },
            density: { value: planetParams.density },
            scale: { value: planetParams.scale },
            lightDirection: { value: planetParams.lightDirection },
            color: { value: planetParams.atmosphereColor },
            pointTexture: { value: cloudTexture.current }
          }}
          transparent={true}
          depthWrite={false}
        />
      </points>
    </group>
  );
}