'use client';

import { useState, useRef, useCallback } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls, Stats } from '@react-three/drei';
import { SubnetLevel, levelNames, generateName, makeIPv6Full } from '@/lib/ipv6names';
import { SubnetTable } from './SubnetTable';
import { SceneController } from './3d/SceneController';

export interface SubnetRow {
  name: string;
  addr: string;
  owner: string;
  percentage: string;
  index: number;
}

interface SpaceNetBrowserProps {
  serverAddr: string;
  httpPort: number;
  playerName: string;
}

export function SpaceNetBrowser({
  serverAddr,
  httpPort,
  playerName
}: SpaceNetBrowserProps) {
  const [currentLevel, setCurrentLevel] = useState<SubnetLevel>(0);
  const [selections, setSelections] = useState<string[]>(Array(8).fill(''));
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [subnets, setSubnets] = useState<SubnetRow[]>([]);
  const [statusMessage, setStatusMessage] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

  const sceneRef = useRef<{ animateForIP: (ip: string) => void }>(null);

  // Generate subnets for current level
  const generateSubnets = useCallback((prefix: string, level: SubnetLevel) => {
    const newSubnets: SubnetRow[] = [];

    // Generate first 100 subnets for performance (in real app, use virtualization)
    for (let i = 0; i < Math.min(1000, 1 << 16); i++) {
      const { addr, subnet } = makeIPv6Full(i, prefix, level);
      const name = generateName(addr, subnet);

      newSubnets.push({
        name,
        addr: `${addr}/${subnet}`,
        owner: '', // Will be fetched from server
        percentage: '', // Will be fetched from server
        index: i
      });
    }

    setSubnets(newSubnets);
  }, []);

  // Handle subnet selection
  const handleSubnetSelect = useCallback((index: number) => {
    setSelectedIndex(index);
    const subnet = subnets[index];

    if (currentLevel < 7) {
      // Navigate deeper
      const addr = subnet.addr.split('/')[0];
      const newPrefix = addr.substring(0, 5 * (currentLevel + 1));
      const newSelections = [...selections];
      newSelections[currentLevel] = newPrefix;
      setSelections(newSelections);

      const newLevel = (currentLevel + 1) as SubnetLevel;
      setCurrentLevel(newLevel);
      generateSubnets(newPrefix, newLevel);
    } else {
      // At deepest level - send claim
      const ip = subnet.addr.split('/')[0];
      sendClaim(ip);
    }

    // Update 3D visualization
    if (sceneRef.current) {
      const ip = subnet.addr.split('/')[0];
      sceneRef.current.animateForIP(ip);
    }
  }, [currentLevel, selections, subnets, generateSubnets]);

  // Handle going back to parent level
  const handleBack = useCallback(() => {
    if (currentLevel > 0) {
      const newLevel = (currentLevel - 1) as SubnetLevel;
      setCurrentLevel(newLevel);

      const parentPrefix = newLevel === 0 ? '' : selections.slice(0, newLevel).join('').substring(0, 5 * newLevel);
      generateSubnets(parentPrefix, newLevel);
    }
  }, [currentLevel, selections, generateSubnets]);

  // Send claim to server
  const sendClaim = async (ip: string) => {
    try {
      // TODO: Implement proof-of-work solving in browser
      // For now, use a placeholder nonce - this will fail validation
      const response = await fetch(`http://[${serverAddr}]:${httpPort}/api/claim/${ip}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          nonce: "0", // TODO: Solve proof-of-work to get valid nonce
          name: playerName
        })
      });

      if (response.status === 201) {
        setStatusMessage('Claim sent successfully!');
        setErrorMessage('');
      } else if (response.status === 422) {
        setErrorMessage('Invalid proof of work - browser PoW not implemented yet');
        setStatusMessage('');
      } else {
        throw new Error(`Server returned status: ${response.status}`);
      }
    } catch (error) {
      setErrorMessage('Failed to send claim: ' + (error as Error).message);
      setStatusMessage('');
    }
  };

  // Initialize with first level
  useState(() => {
    generateSubnets('', 0);
  });

  return (
    <div className="h-screen bg-black text-white p-4">
      <div className="mb-4">
        <h1 className="text-2xl font-bold mb-2">SpaceNet Browser</h1>
        <div className="text-sm text-gray-400">
          Level: {levelNames[currentLevel]} ({currentLevel + 1}/8)
        </div>
      </div>

      <div className="flex h-[calc(100vh-120px)] gap-4">
        {/* Left panel - Subnet table */}
        <div className="w-1/2">
          <SubnetTable
            subnets={subnets}
            selectedIndex={selectedIndex}
            onSelect={handleSubnetSelect}
            onIndexChange={setSelectedIndex}
            onBack={handleBack}
            canGoBack={currentLevel > 0}
          />
        </div>

        {/* Right panel - 3D visualization */}
        <div className="w-1/2 border border-gray-700 rounded">
          <Canvas
            camera={{ position: [0, 0, 10] }}
          >
            <SceneController
              ref={sceneRef}
              level={currentLevel}
              selectedIP={subnets[selectedIndex]?.addr.split('/')[0] || '::'}
            />
            <OrbitControls
              enablePan={true}
              enableZoom={true}
              enableRotate={true}
            />
            <Stats />
          </Canvas>
        </div>
      </div>

      {/* Status messages */}
      <div className="mt-4 h-8">
        {statusMessage && (
          <div className="text-green-400">{statusMessage}</div>
        )}
        {errorMessage && (
          <div className="text-red-400">{errorMessage}</div>
        )}
        <div className="text-gray-500 text-sm">
          Controls: Click to select subnet, Esc to go back, Mouse to navigate 3D view
        </div>
      </div>
    </div>
  );
}