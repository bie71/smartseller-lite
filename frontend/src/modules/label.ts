import { jsPDF } from 'jspdf';
import html2canvas from 'html2canvas';
import { createApp, h } from 'vue';
import LabelPreview from '../ui/components/LabelPreview.vue';
import type { AppSettings } from './settings';

export type LabelData = {
  id: string;
  senderName: string;
  senderPhone?: string;
  senderAddress?: string;
  recipientName: string;
  recipientPhone?: string;
  recipientAddress?: string;
  courier?: string;
  service?: string;
  trackingCode?: string;
  orderCode?: string;
  notes?: string;
  isCOD?: boolean;
  codAmount?: number;
  weight?: number; // in kg
};

const MM_PER_INCH = 25.4;
const DPI = 96;
const PX_TO_MM = (px: number) => (px / DPI) * MM_PER_INCH;
const CANVAS_SCALE = 2;

async function renderComponentToCanvas(label: LabelData, widthPx: number, appSettings: AppSettings | null): Promise<HTMLCanvasElement> {
  return new Promise((resolve, reject) => {
    const container = document.createElement('div');
    container.style.position = 'absolute';
    container.style.left = '-9999px';
    container.style.top = '-9999px';
    container.style.width = `${widthPx}px`;
    document.body.appendChild(container);

    const app = createApp({
      render() {
        return h(LabelPreview, { 
          label, 
          isPrinting: true, 
          appSettings 
        });
      },
    });

    app.mount(container);

    setTimeout(async () => {
      try {
        const el = container.querySelector<HTMLElement>(`#label-preview-${label.id}`);
        if (!el) throw new Error('Rendered element not found');
        
        const canvas = await html2canvas(el, {
          scale: CANVAS_SCALE,
          useCORS: true,
          logging: false,
          backgroundColor: null, // Use transparent background
        });
        document.body.removeChild(container);
        app.unmount();
        resolve(canvas);
      } catch (error) {
        try {
          document.body.removeChild(container);
          app.unmount();
        } catch (e) {}
        reject(error);
      }
    }, 200);
  });
}

export async function generateSingleLabelPdf(label: LabelData, appSettings: AppSettings | null): Promise<Blob> {
  const doc = new jsPDF({ unit: 'mm', format: 'a4' });
  const pageWidth = doc.internal.pageSize.getWidth();
  const margin = 10;
  const labelWidthMm = pageWidth - (margin * 2);
  const labelWidthPx = (labelWidthMm / MM_PER_INCH) * DPI;

  const canvas = await renderComponentToCanvas(label, labelWidthPx, appSettings);
  const imgData = canvas.toDataURL('image/png');

  const finalWidthMm = PX_TO_MM(canvas.width / CANVAS_SCALE);
  const finalHeightMm = PX_TO_MM(canvas.height / CANVAS_SCALE);

  const x = (pageWidth - finalWidthMm) / 2;
  const y = margin;

  doc.addImage(imgData, 'PNG', x, y, finalWidthMm, finalHeightMm);
  return doc.output('blob');
}

export async function generateLabelsPdf(labels: LabelData[], appSettings: AppSettings | null): Promise<Blob> {
  if (!labels.length) throw new Error('Tidak ada label untuk dicetak.');

  const doc = new jsPDF({ unit: 'mm', format: 'a4' });
  const pageWidth = doc.internal.pageSize.getWidth();
  const margin = 10;
  const labelWidthMm = pageWidth - (margin * 2);
  const labelWidthPx = (labelWidthMm / MM_PER_INCH) * DPI;

  for (let i = 0; i < labels.length; i++) {
    if (i > 0) {
      doc.addPage();
    }

    const label = labels[i];
    const canvas = await renderComponentToCanvas(label, labelWidthPx, appSettings);
    const imgData = canvas.toDataURL('image/png');

    const finalWidthMm = PX_TO_MM(canvas.width / CANVAS_SCALE);
    const finalHeightMm = PX_TO_MM(canvas.height / CANVAS_SCALE);

    const x = (pageWidth - finalWidthMm) / 2;
    const y = margin;

    doc.addImage(imgData, 'PNG', x, y, finalWidthMm, finalHeightMm);
  }

  return doc.output('blob');
}