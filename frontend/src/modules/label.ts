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
const CANVAS_SCALE = 2;

let html2canvasLoader: Promise<typeof import('html2canvas')> | null = null;
let jsPdfLoader: Promise<typeof import('jspdf')> | null = null;

async function ensureHtml2Canvas() {
  if (!html2canvasLoader) {
    html2canvasLoader = import('html2canvas');
  }
  return html2canvasLoader;
}

async function ensureJsPdf() {
  if (!jsPdfLoader) {
    jsPdfLoader = import('jspdf');
  }
  return jsPdfLoader;
}


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
        
        const { default: html2canvas } = await ensureHtml2Canvas();
        const canvas = await html2canvas(el, {
          scale: CANVAS_SCALE,
          useCORS: true,
          logging: false,
          backgroundColor: '#ffffff',
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
  const { jsPDF } = await ensureJsPdf();
  const doc = new jsPDF({ unit: 'mm', format: 'a4' });
  const pageWidth = doc.internal.pageSize.getWidth();
  const margin = 10;
  const labelWidthMm = pageWidth - (margin * 2);
  const labelWidthPx = (labelWidthMm / MM_PER_INCH) * DPI;

  const canvas = await renderComponentToCanvas(label, labelWidthPx, appSettings);
  const imgData = canvas.toDataURL('image/png');

  const aspectRatio = canvas.height / canvas.width;
  const finalWidthMm = labelWidthMm;
  const finalHeightMm = finalWidthMm * aspectRatio;

  const x = margin;
  const y = margin;

  doc.addImage(imgData, 'PNG', x, y, finalWidthMm, finalHeightMm);
  return doc.output('blob');
}

export async function generateLabelsPdf(labels: LabelData[], appSettings: AppSettings | null): Promise<Blob> {
  if (!labels.length) throw new Error('Tidak ada label untuk dicetak.');

  const { jsPDF } = await ensureJsPdf();
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

    const aspectRatio = canvas.height / canvas.width;
    const finalWidthMm = labelWidthMm;
    const finalHeightMm = finalWidthMm * aspectRatio;

    const x = margin;
    const y = margin;

    doc.addImage(imgData, 'PNG', x, y, finalWidthMm, finalHeightMm);
  }

  return doc.output('blob');
}