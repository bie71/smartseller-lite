import React, { useState, useEffect } from 'react';
import { LabelData } from '../types';
import { Icon } from './Icon';

interface OrderFormProps {
  onAddOrder: (order: Omit<LabelData, 'id'>) => void;
  onUpdateOrder: (order: LabelData) => void;
  onCancelEdit: () => void;
  editingLabel: LabelData | null;
}

const generateNewInitialData = (): Omit<LabelData, 'id'> => ({
  senderName: 'Toko SmartSeller',
  senderPhone: '081234567890',
  senderAddress: 'Jl. Teknologi No. 1, Jakarta',
  recipientName: 'Budi Santoso',
  recipientPhone: '089876543210',
  recipientAddress: 'Jl. Pahlawan No. 45, Apartemen Sejahtera Lt. 10, Surabaya, 60241',
  orderId: `INV-${new Date().getFullYear()}${(new Date().getMonth() + 1).toString().padStart(2, '0')}${(new Date().getDate()).toString().padStart(2, '0')}-${Math.floor(1000 + Math.random() * 9000)}`,
  trackingNumber: `SS${Date.now()}`,
  items: '1x Kemeja Lengan Panjang, 2x Celana Chino',
  courier: 'JNE REG',
  isCOD: true,
  codAmount: 250000,
  notes: 'Mohon taruh di resepsionis jika tidak ada orang.',
  weight: 1.2
});

const InputField: React.FC<React.InputHTMLAttributes<HTMLInputElement> & { label: string; error?: string }> = ({ label, id, error, ...props }) => (
  <div>
    <label htmlFor={id} className="block text-sm font-medium text-slate-700 mb-1">{label}</label>
    <input 
      id={id} 
      {...props} 
      className={`block w-full px-3 py-2 bg-white border rounded-md text-sm shadow-sm placeholder-slate-400 focus:outline-none focus:ring-1 transition-colors duration-200 
        ${error 
          ? 'border-red-500 text-red-900 placeholder-red-300 focus:ring-red-500 focus:border-red-500' 
          : 'border-slate-300 focus:ring-blue-500 focus:border-blue-500 hover:border-slate-400'
        }`}
      aria-invalid={!!error}
      aria-describedby={error ? `${id}-error` : undefined}
    />
    {error && <p id={`${id}-error`} className="mt-1 text-xs text-red-600">{error}</p>}
  </div>
);

const TextAreaField: React.FC<React.TextareaHTMLAttributes<HTMLTextAreaElement> & { label: string; error?: string }> = ({ label, id, error, ...props }) => (
    <div>
      <label htmlFor={id} className="block text-sm font-medium text-slate-700 mb-1">{label}</label>
      <textarea 
        id={id} 
        {...props} 
        rows={3} 
        className={`block w-full px-3 py-2 bg-white border rounded-md text-sm shadow-sm placeholder-slate-400 focus:outline-none focus:ring-1 transition-colors duration-200 
        ${error 
          ? 'border-red-500 text-red-900 placeholder-red-300 focus:ring-red-500 focus:border-red-500' 
          : 'border-slate-300 focus:ring-blue-500 focus:border-blue-500 hover:border-slate-400'
        }`}
        aria-invalid={!!error}
        aria-describedby={error ? `${id}-error` : undefined}
      />
      {error && <p id={`${id}-error`} className="mt-1 text-xs text-red-600">{error}</p>}
    </div>
);


export const OrderForm: React.FC<OrderFormProps> = ({ onAddOrder, onUpdateOrder, onCancelEdit, editingLabel }) => {
  const [formData, setFormData] = useState(generateNewInitialData());
  const [errors, setErrors] = useState<Partial<Record<keyof Omit<LabelData, 'id'>, string>>>({});
  const isEditing = !!editingLabel;

  useEffect(() => {
    if (editingLabel) {
      setFormData(editingLabel);
    } else {
      setFormData(generateNewInitialData());
    }
    setErrors({}); // Reset errors when editing label changes
  }, [editingLabel]);

  const validateForm = (data: Omit<LabelData, 'id'>): Partial<Record<keyof Omit<LabelData, 'id'>, string>> => {
    const newErrors: Partial<Record<keyof Omit<LabelData, 'id'>, string>> = {};

    if (!data.senderName.trim()) newErrors.senderName = 'Nama Pengirim wajib diisi.';
    if (!data.senderPhone.trim()) newErrors.senderPhone = 'No. Telepon Pengirim wajib diisi.';
    if (!data.recipientName.trim()) newErrors.recipientName = 'Nama Penerima wajib diisi.';
    if (!data.recipientPhone.trim()) newErrors.recipientPhone = 'No. Telepon Penerima wajib diisi.';
    if (!data.recipientAddress.trim()) newErrors.recipientAddress = 'Alamat Lengkap Penerima wajib diisi.';
    if (!data.orderId.trim()) newErrors.orderId = 'ID Pesanan wajib diisi.';
    if (!data.trackingNumber.trim()) newErrors.trackingNumber = 'No. Resi wajib diisi.';
    if (!data.items.trim()) newErrors.items = 'Isi Paket wajib diisi.';
    if (!data.courier.trim()) newErrors.courier = 'Kurir Pengiriman wajib diisi.';

    const weightNum = Number(data.weight);
    if (isNaN(weightNum) || weightNum <= 0) {
        newErrors.weight = 'Berat harus berupa angka lebih dari 0.';
    }

    if (data.isCOD) {
        const codNum = Number(data.codAmount);
        if (isNaN(codNum) || codNum < 0) {
            newErrors.codAmount = 'Jumlah COD harus berupa angka yang valid.';
        }
    }
    return newErrors;
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target;
    
    // Clear error for the field being changed
    if (errors[name as keyof typeof errors]) {
        setErrors(prev => {
            const newErrors = { ...prev };
            delete newErrors[name as keyof typeof errors];
            return newErrors;
        });
    }
    
    if (name === 'codAmount') {
      const numericValue = value.replace(/[^0-9]/g, '');
      const number = parseInt(numericValue, 10);
      setFormData(prev => ({ ...prev, codAmount: isNaN(number) ? 0 : number }));
    } else if (type === 'checkbox' && e.target instanceof HTMLInputElement) {
        setFormData(prev => ({ ...prev, [name]: e.target.checked }));
    } else {
        setFormData(prev => ({ ...prev, [name]: value }));
    }
  };

  const handleBlur = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name } = e.target as { name: keyof LabelData };
    const formErrors = validateForm(formData);
    if (formErrors[name]) {
      setErrors(prev => ({ ...prev, [name]: formErrors[name] }));
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const formErrors = validateForm(formData);
    if (Object.keys(formErrors).length > 0) {
        setErrors(formErrors);
        return;
    }

    const dataToSubmit = {
      ...formData,
      codAmount: Number(formData.codAmount),
      weight: Number(formData.weight)
    }
    if (isEditing) {
      onUpdateOrder(dataToSubmit as LabelData);
    } else {
      onAddOrder(dataToSubmit);
      setFormData(generateNewInitialData());
      setErrors({}); // Clear errors for new form
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6 bg-white p-6 rounded-lg shadow">
      <h2 className="text-xl font-semibold text-slate-800 border-b pb-3">{isEditing ? 'Edit Detail Pengiriman' : 'Detail Pengiriman'}</h2>
      
      {/* Sender & Recipient */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-4">
          <h3 className="font-medium text-slate-600">Pengirim</h3>
          <InputField label="Nama Pengirim" id="senderName" name="senderName" value={formData.senderName} onChange={handleChange} onBlur={handleBlur} error={errors.senderName} required />
          <InputField label="No. Telepon" id="senderPhone" name="senderPhone" value={formData.senderPhone} onChange={handleChange} onBlur={handleBlur} error={errors.senderPhone} required />
        </div>
        <div className="space-y-4">
          <h3 className="font-medium text-slate-600">Penerima</h3>
          <InputField label="Nama Penerima" id="recipientName" name="recipientName" value={formData.recipientName} onChange={handleChange} onBlur={handleBlur} error={errors.recipientName} required />
          <InputField label="No. Telepon" id="recipientPhone" name="recipientPhone" value={formData.recipientPhone} onChange={handleChange} onBlur={handleBlur} error={errors.recipientPhone} required />
          <TextAreaField label="Alamat Lengkap" id="recipientAddress" name="recipientAddress" value={formData.recipientAddress} onChange={handleChange} onBlur={handleBlur} error={errors.recipientAddress} required />
        </div>
      </div>

      {/* Order Details */}
      <div className="space-y-4 border-t pt-6">
        <h3 className="font-medium text-slate-600">Detail Pesanan</h3>
        <InputField label="ID Pesanan" id="orderId" name="orderId" value={formData.orderId} onChange={handleChange} onBlur={handleBlur} error={errors.orderId} required />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <InputField label="No. Resi" id="trackingNumber" name="trackingNumber" value={formData.trackingNumber} onChange={handleChange} onBlur={handleBlur} error={errors.trackingNumber} required />
            <InputField label="Kurir Pengiriman" id="courier" name="courier" value={formData.courier} onChange={handleChange} onBlur={handleBlur} error={errors.courier} required />
        </div>
        <InputField label="Berat (kg)" id="weight" name="weight" type="number" step="0.1" value={formData.weight} onChange={handleChange} onBlur={handleBlur} error={errors.weight} required />
        <TextAreaField label="Isi Paket" id="items" name="items" value={formData.items} onChange={handleChange} onBlur={handleBlur} error={errors.items} required />
        <TextAreaField label="Catatan (Opsional)" id="notes" name="notes" value={formData.notes} onChange={handleChange} />
      </div>

      {/* COD */}
      <div className="space-y-4 border-t pt-6">
        <div className="flex items-start">
            <div className="flex items-center h-5">
                <input id="isCOD" name="isCOD" type="checkbox" checked={formData.isCOD} onChange={handleChange} className="focus:ring-blue-500 h-4 w-4 text-blue-600 border-slate-300 rounded" />
            </div>
            <div className="ml-3 text-sm">
                <label htmlFor="isCOD" className="font-medium text-slate-700">Cash on Delivery (COD)</label>
                <p className="text-slate-500">Aktifkan jika pesanan ini menggunakan metode COD.</p>
            </div>
        </div>
        {formData.isCOD && (
            <InputField 
              label="Jumlah COD (IDR)" 
              id="codAmount" 
              name="codAmount" 
              type="text" 
              inputMode="numeric"
              value={new Intl.NumberFormat('id-ID').format(formData.codAmount || 0)} 
              onChange={handleChange} 
              onBlur={handleBlur} 
              error={errors.codAmount} 
              required={formData.isCOD} 
            />
        )}
      </div>

      <div className="pt-4 border-t space-y-3">
        <button type="submit" className="w-full bg-slate-800 hover:bg-slate-700 text-white font-bold py-3 px-4 rounded-lg flex items-center justify-center transition-colors">
          <Icon name={isEditing ? 'pencil' : 'package'} className="w-5 h-5 mr-2"/>
          {isEditing ? 'Update Label' : 'Tambahkan Label ke Antrian'}
        </button>
        {isEditing && (
            <button type="button" onClick={onCancelEdit} className="w-full bg-slate-200 hover:bg-slate-300 text-slate-800 font-bold py-3 px-4 rounded-lg flex items-center justify-center transition-colors">
                <Icon name="x" className="w-5 h-5 mr-2"/>
                Batal Edit
            </button>
        )}
      </div>
    </form>
  );
};