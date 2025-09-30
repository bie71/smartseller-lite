import React from 'react';
import { LabelData } from '../types';
import { Icon } from './Icon';

interface LabelPreviewProps {
  label: LabelData;
  onRemove: (id: string) => void;
  onEdit: (id: string) => void;
}

const BarcodePlaceholder: React.FC<{ trackingId: string }> = ({ trackingId }) => (
  <div>
    <svg width="100%" height="40" className="text-black">
      <rect x="0" y="5" width="2" height="30" fill="currentColor" />
      <rect x="4" y="5" width="4" height="30" fill="currentColor" />
      <rect x="10" y="5" width="2" height="30" fill="currentColor" />
      <rect x="14" y="5" width="6" height="30" fill="currentColor" />
      <rect x="22" y="5" width="2" height="30" fill="currentColor" />
      <rect x="28" y="5" width="4" height="30" fill="currentColor" />
      <rect x="34" y="5" width="2" height="30" fill="currentColor" />
      <rect x="38" y="5" width="2" height="30" fill="currentColor" />
      <rect x="44" y="5" width="4" height="30" fill="currentColor" />
      <rect x="52" y="5" width="2" height="30" fill="currentColor" />
      <rect x="56" y="5" width="4" height="30" fill="currentColor" />
      <rect x="64" y="5" width="2" height="30" fill="currentColor" />
      <rect x="70" y="5" width="6" height="30" fill="currentColor" />
      <rect x="78" y="5" width="2" height="30" fill="currentColor" />
      <rect x="82" y="5" width="4" height="30" fill="currentColor" />
      <rect x="90" y="5" width="2" height="30" fill="currentColor" />
      <rect x="94" y="5" width="2" height="30" fill="currentColor" />
      <rect x="100" y="5" width="4" height="30" fill="currentColor" />
      <rect x="106" y="5" width="2" height="30" fill="currentColor" />
      <rect x="110" y="5" width="2" height="30" fill="currentColor" />
      <rect x="116" y="5" width="6" height="30" fill="currentColor" />
      <text x="50%" y="4" textAnchor="middle" fontSize="8" fontFamily="monospace" fill="currentColor">RESI PENGIRIMAN</text>
    </svg>
    <p className="text-center font-mono font-bold tracking-widest text-sm mt-1">{trackingId}</p>
  </div>
);


export const LabelPreview: React.FC<LabelPreviewProps> = ({ label, onRemove, onEdit }) => {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(amount);
  };
    
  return (
    <div id={`label-preview-${label.id}`} className="bg-white p-4 border border-slate-300 rounded-lg shadow-sm w-full relative font-sans text-xs text-slate-900 leading-tight flex flex-col">
      <div className="absolute top-2 right-2 flex space-x-1 z-10 print-hidden">
        <button 
          onClick={() => onEdit(label.id)}
          className="p-1 bg-white/50 backdrop-blur-sm text-blue-500 hover:bg-blue-100 hover:text-blue-700 rounded-full transition-colors"
          aria-label="Edit label"
        >
          <Icon name="pencil" className="w-4 h-4" />
        </button>
        <button 
          onClick={() => onRemove(label.id)}
          className="p-1 bg-white/50 backdrop-blur-sm text-red-500 hover:bg-red-100 hover:text-red-700 rounded-full transition-colors"
          aria-label="Remove label"
        >
          <Icon name="x" className="w-4 h-4" />
        </button>
      </div>

      {/* Header */}
      <div className="flex justify-between items-start pb-2 border-b-2 border-black flex-shrink-0">
        <div className="flex-1">
          <p className="font-bold text-lg uppercase">{label.courier}</p>
          <p className="text-slate-600">Berat: {label.weight} kg</p>
        </div>
        <div className="text-right">
          <p className="font-bold text-sm">SmartSeller</p>
          {label.isCOD && (
            <div className="mt-1 px-2 py-1 bg-black text-white font-bold text-sm rounded">
              COD: {formatCurrency(label.codAmount)}
            </div>
          )}
        </div>
      </div>
      
      {/* Main Content Area: Stretches to fill space */}
      <div className="flex-grow min-h-0">
        {/* Addresses */}
        <div className="grid grid-cols-2 gap-2 py-2 border-b border-dashed border-slate-400">
          <div className="space-y-2">
            <div>
              <p className="font-semibold text-slate-500">PENGIRIM:</p>
              <p className="font-bold">{label.senderName}</p>
              <p className="break-words">{label.senderPhone}</p>
            </div>
            <div>
              <p className="font-semibold text-slate-500">ID PESANAN:</p>
              <p className="font-mono tracking-wider">{label.orderId}</p>
            </div>
          </div>
          <div>
            <p className="font-semibold text-slate-500">PENERIMA:</p>
            <p className="font-bold text-base">{label.recipientName}</p>
            <p className="break-words">{label.recipientPhone}</p>
            <p className="mt-1 break-words">{label.recipientAddress}</p>
          </div>
        </div>

        {/* Order Details */}
        <div className="py-2 space-y-2">
         <div>
            <p className="font-semibold text-slate-500">BARANG:</p>
            <p className="font-medium break-words">{label.items}</p>
        </div>
        {label.notes && (
            <div>
                <p className="font-semibold text-slate-500">CATATAN:</p>
                <p className="font-medium break-words">{label.notes}</p>
            </div>
        )}
      </div>
      </div>


      {/* Barcode and Order ID (Footer): Pushed to the bottom */}
      <div className="mt-auto pt-2 border-t border-dashed border-slate-400 flex-shrink-0">
        <BarcodePlaceholder trackingId={label.trackingNumber} />
      </div>
    </div>
  );
};