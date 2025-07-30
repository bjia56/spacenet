'use client';

import { useEffect, useRef } from 'react';
import { SubnetRow } from './SpaceNetBrowser';

interface SubnetTableProps {
  subnets: SubnetRow[];
  selectedIndex: number;
  onSelect: (index: number) => void;
  onIndexChange: (index: number) => void;
  onBack: () => void;
  canGoBack: boolean;
}

export function SubnetTable({ 
  subnets, 
  selectedIndex, 
  onSelect, 
  onIndexChange,
  onBack, 
  canGoBack 
}: SubnetTableProps) {
  const tableRef = useRef<HTMLDivElement>(null);
  const selectedRowRef = useRef<HTMLDivElement>(null);

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowUp':
          e.preventDefault();
          if (selectedIndex > 0) {
            onIndexChange(selectedIndex - 1);
          }
          break;
        case 'ArrowDown':
          e.preventDefault();
          if (selectedIndex < subnets.length - 1) {
            onIndexChange(selectedIndex + 1);
          }
          break;
        case 'Enter':
          e.preventDefault();
          onSelect(selectedIndex);
          break;
        case 'Escape':
          e.preventDefault();
          if (canGoBack) {
            onBack();
          }
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedIndex, subnets.length, onSelect, onIndexChange, onBack, canGoBack]);

  // Auto-scroll to selected row
  useEffect(() => {
    if (selectedRowRef.current && tableRef.current) {
      selectedRowRef.current.scrollIntoView({
        behavior: 'smooth',
        block: 'nearest'
      });
    }
  }, [selectedIndex]);

  return (
    <div className="border border-gray-700 rounded p-2 h-full flex flex-col">
      <div className="mb-2">
        <div className="grid grid-cols-12 gap-2 text-sm font-bold text-gray-300 border-b border-gray-700 pb-2">
          <div className="col-span-6">Subnet</div>
          <div className="col-span-3">Owner</div>
          <div className="col-span-3">Percentage</div>
        </div>
      </div>
      
      <div 
        ref={tableRef}
        className="flex-1 overflow-y-auto space-y-1"
      >
        {subnets.map((subnet, index) => (
          <div
            key={`${subnet.addr}-${index}`}
            ref={index === selectedIndex ? selectedRowRef : null}
            className={`
              grid grid-cols-12 gap-2 p-2 text-sm cursor-pointer rounded
              transition-colors duration-150
              ${index === selectedIndex 
                ? 'bg-blue-600 text-white' 
                : 'hover:bg-gray-800 text-gray-300'
              }
            `}
            onClick={() => onSelect(index)}
          >
            <div className="col-span-6 truncate font-mono text-xs">
              {subnet.name}
            </div>
            <div className="col-span-3 truncate">
              {subnet.owner || '-'}
            </div>
            <div className="col-span-3 truncate">
              {subnet.percentage || '-'}
            </div>
          </div>
        ))}
      </div>

      <div className="mt-2 pt-2 border-t border-gray-700 text-xs text-gray-500">
        Showing {subnets.length.toLocaleString()} subnets
        {canGoBack && (
          <button
            onClick={onBack}
            className="ml-4 px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-white"
          >
            ← Back
          </button>
        )}
      </div>
    </div>
  );
}