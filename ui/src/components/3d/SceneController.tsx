'use client';

import { forwardRef, useImperativeHandle, useMemo } from 'react';
import { SubnetLevel } from '@/lib/ipv6names';
import { StarrySkybox } from './StarrySkybox';
import { GreatWall3D } from './GreatWall3D';
import { Supercluster3D } from './Supercluster3D';
import { GalaxyGroup3D } from './GalaxyGroup3D';
import { Galaxy3D } from './Galaxy3D';
import { StarCluster3D } from './StarCluster3D';
import { SolarSystem3D } from './SolarSystem3D';
import { Planet3D } from './Planet3D';
import { City3D } from './City3D';

interface SceneControllerProps {
  level: SubnetLevel;
  selectedIP: string;
}

export interface SceneControllerRef {
  animateForIP: (ip: string) => void;
}

export const SceneController = forwardRef<SceneControllerRef, SceneControllerProps>(
  ({ level, selectedIP }, ref) => {
    // Create a seed from the IP address for deterministic randomization
    const ipSeed = useMemo(() => {
      if (!selectedIP) return 0;
      let hash = 0;
      for (let i = 0; i < selectedIP.length; i++) {
        const char = selectedIP.charCodeAt(i);
        hash = ((hash << 5) - hash) + char;
        hash = hash & hash; // Convert to 32bit integer
      }
      return Math.abs(hash);
    }, [selectedIP]);

    useImperativeHandle(ref, () => ({
      animateForIP: (ip: string) => {
        // The IP change will trigger a re-render with new ipSeed
        // Individual components will handle their own animation updates
      }
    }));

    // Render the appropriate 3D visualization based on current level
    const renderVisualization = () => {
      const commonProps = { ipSeed };

      switch (level) {
        case 0: // Great Wall (/16)
          return <GreatWall3D {...commonProps} />;
        case 1: // Supercluster (/32)
          return <Supercluster3D {...commonProps} />;
        case 2: // Galaxy Group (/48)
          return <GalaxyGroup3D {...commonProps} />;
        case 3: // Galaxy (/64)
          return <Galaxy3D {...commonProps} />;
        case 4: // Star Cluster (/80)
          return <StarCluster3D {...commonProps} />;
        case 5: // Solar System (/96)
          return <SolarSystem3D {...commonProps} />;
        case 6: // Planet (/112)
          return <Planet3D {...commonProps} />;
        case 7: // City (/128)
          return <City3D {...commonProps} />;
        default:
          return <GreatWall3D {...commonProps} />;
      }
    };

    return (
      <group>
        <StarrySkybox seed={ipSeed} />
        {renderVisualization()}
      </group>
    );
  }
);

SceneController.displayName = 'SceneController';