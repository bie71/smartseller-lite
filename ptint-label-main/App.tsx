import React, { useState, useCallback, useEffect, useRef } from 'react';
import { LabelData } from './types';
import { Header } from './components/Header';
import { OrderForm } from './components/OrderForm';
import { LabelPreview } from './components/LabelPreview';
import { Icon } from './components/Icon';
import { generatePdf } from './services/pdfService';

const APP_STORAGE_KEY = 'smart-seller-labels';

const App: React.FC = () => {
  const [labels, setLabels] = useState<LabelData[]>(() => {
    try {
      const savedLabels = window.localStorage.getItem(APP_STORAGE_KEY);
      return savedLabels ? JSON.parse(savedLabels) : [];
    } catch (error) {
      console.error("Error reading labels from localStorage:", error);
      alert("Gagal memuat label dari penyimpanan lokal. Mungkin penyimpanan Anda penuh atau rusak.");
      return [];
    }
  });
  
  const [isPrinting, setIsPrinting] = useState(false);
  const [editingLabel, setEditingLabel] = useState<LabelData | null>(null);
  const [confirmingClearAll, setConfirmingClearAll] = useState(false);
  const confirmTimeoutRef = useRef<number | null>(null);

  useEffect(() => {
    try {
      window.localStorage.setItem(APP_STORAGE_KEY, JSON.stringify(labels));
    } catch (error) {
      console.error("Error saving labels to localStorage:", error);
      alert("Gagal menyimpan label ke penyimpanan lokal. Perubahan terbaru mungkin tidak akan tersimpan.");
    }
  }, [labels]);

  useEffect(() => {
    // Cleanup timeout on component unmount
    return () => {
      if (confirmTimeoutRef.current) {
        clearTimeout(confirmTimeoutRef.current);
      }
    };
  }, []);

  const handleAddOrder = useCallback((order: Omit<LabelData, 'id'>) => {
    const newLabel = {
      ...order,
      id: `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
    };
    setLabels(prevLabels => [...prevLabels, newLabel]);
  }, []);

  const handleUpdateOrder = useCallback((updatedLabel: LabelData) => {
    setLabels(prevLabels => prevLabels.map(label => label.id === updatedLabel.id ? updatedLabel : label));
    setEditingLabel(null);
  }, []);

  const handleStartEdit = useCallback((id: string) => {
    const labelToEdit = labels.find(label => label.id === id);
    if (labelToEdit) {
      setEditingLabel(labelToEdit);
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  }, [labels]);
  
  const handleCancelEdit = useCallback(() => {
    setEditingLabel(null);
  }, []);

  const handleRemoveLabel = useCallback((id: string) => {
    setLabels(prevLabels => prevLabels.filter(label => label.id !== id));
  }, []);

  const handleClearAllClick = useCallback(() => {
    if (confirmingClearAll) {
      setLabels([]);
      setConfirmingClearAll(false);
      if (confirmTimeoutRef.current) {
        clearTimeout(confirmTimeoutRef.current);
        confirmTimeoutRef.current = null;
      }
    } else {
      setConfirmingClearAll(true);
      confirmTimeoutRef.current = window.setTimeout(() => {
        setConfirmingClearAll(false);
      }, 3000); // 3-second window to confirm
    }
  }, [confirmingClearAll]);

  const handlePrint = async () => {
    if (labels.length === 0) {
      alert("Tidak ada label untuk dicetak. Silakan tambahkan label terlebih dahulu.");
      return;
    }
    setIsPrinting(true);
    try {
      await generatePdf(labels);
    } catch (error) {
      console.error("Failed to generate PDF:", error);
      alert("Gagal membuat PDF. Silakan coba lagi.");
    } finally {
      setIsPrinting(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-100">
      <Header />
      <main className="container mx-auto p-4 sm:p-6 lg:p-8">
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-8">
          
          {/* Left Column: Form */}
          <div className="lg:col-span-2">
            <OrderForm 
              onAddOrder={handleAddOrder}
              onUpdateOrder={handleUpdateOrder}
              editingLabel={editingLabel}
              onCancelEdit={handleCancelEdit}
            />
          </div>

          {/* Right Column: Preview Queue */}
          <div className="lg:col-span-3">
            <div className="bg-white p-6 rounded-lg shadow sticky top-8">
              <div className="flex justify-between items-center border-b pb-3 mb-4">
                <h2 className="text-xl font-semibold text-slate-800">Antrian Cetak ({labels.length})</h2>
                <div className="flex items-center space-x-2">
                  {labels.length > 0 && (
                     <button
                        onClick={handleClearAllClick}
                        disabled={isPrinting}
                        className={`font-bold py-2 rounded-lg flex items-center justify-center transition-all duration-300 ease-in-out ${
                            confirmingClearAll
                            ? 'bg-yellow-400 hover:bg-yellow-500 text-slate-800 w-24 px-3'
                            : 'bg-red-500 hover:bg-red-600 text-white w-12 px-3'
                        } disabled:bg-slate-300 disabled:cursor-not-allowed`}
                        aria-label={confirmingClearAll ? "Konfirmasi hapus semua" : "Hapus semua label"}
                    >
                        {confirmingClearAll ? (
                            'Yakin?'
                        ) : (
                            <Icon name="trash" className="w-5 h-5" />
                        )}
                    </button>
                  )}
                  <button
                    onClick={handlePrint}
                    disabled={isPrinting || labels.length === 0}
                    className="bg-blue-600 hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed text-white font-bold py-2 px-4 rounded-lg flex items-center transition-colors"
                  >
                    {isPrinting ? (
                      <Icon name="spinner" className="w-5 h-5 mr-2 animate-spin" />
                    ) : (
                      <Icon name="printer" className="w-5 h-5 mr-2" />
                    )}
                    {isPrinting ? 'Mencetak...' : `Cetak Semua Label`}
                  </button>
                </div>
              </div>

              {labels.length > 0 ? (
                <div className="max-h-[80vh] overflow-y-auto pr-2 -mr-2 space-y-4 grid grid-cols-1 xl:grid-cols-2 gap-4">
                  {labels.map(label => (
                    <LabelPreview 
                      key={label.id} 
                      label={label} 
                      onRemove={handleRemoveLabel}
                      onEdit={handleStartEdit}
                    />
                  ))}
                </div>
              ) : (
                <div className="text-center py-16 px-6 border-2 border-dashed border-slate-300 rounded-lg">
                  <Icon name="package" className="mx-auto h-12 w-12 text-slate-400" />
                  <h3 className="mt-2 text-sm font-medium text-slate-900">Antrian Cetak Kosong</h3>
                  <p className="mt-1 text-sm text-slate-500">Isi formulir di samping untuk menambahkan label.</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default App;