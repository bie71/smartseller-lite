
import React from 'react';
import { Icon } from './Icon';

export const Header: React.FC = () => {
  return (
    <header className="bg-white shadow-md">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center space-x-3">
            <div className="bg-slate-800 text-white p-2 rounded-lg">
              <Icon name="package" className="w-6 h-6" />
            </div>
            <h1 className="text-2xl font-bold text-slate-800">
              SmartSeller <span className="font-light">Label Printer</span>
            </h1>
          </div>
        </div>
      </div>
    </header>
  );
};
