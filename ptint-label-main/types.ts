export interface LabelData {
  id: string;
  senderName: string;
  senderPhone: string;
  senderAddress: string;
  recipientName: string;
  recipientPhone: string;
  recipientAddress: string;
  orderId: string;
  trackingNumber: string;
  items: string;
  courier: string;
  isCOD: boolean;
  codAmount: number;
  notes: string;
  weight: number;
}