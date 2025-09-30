import { jsPDF } from 'jspdf';
import html2canvas from 'html2canvas';
import { createApp, h } from 'vue';
import LabelPreview from '../ui/components/LabelPreview.vue';

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
const CANVAS_SCALE = 2; // 2 is a good balance of quality and performance

async function renderComponentToCanvas(label: LabelData, widthPx: number): Promise<HTMLCanvasElement> {
  return new Promise((resolve, reject) => {
    const container = document.createElement('div');
    container.style.position = 'absolute';
    container.style.left = '-9999px';
    container.style.top = '-9999px';
    container.style.width = `${widthPx}px`;
    document.body.appendChild(container);

    const app = createApp({
      render() {
        return h(LabelPreview, { label });
      },
    });

    app.mount(container);

    setTimeout(async () => {
      try {
        const el = container.querySelector<HTMLElement>(`#label-preview-${label.id}`);
        if (!el) {
          throw new Error('Rendered element not found');
        }
        const canvas = await html2canvas(el, {
          scale: CANVAS_SCALE,
          useCORS: true,
          logging: false,
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
    }, 200); // Increased timeout for potentially complex renders
  });
}

export async function generateSingleLabelPdf(label: LabelData): Promise<Blob> {
  const labelWidthMm = 100;
  const labelWidthPx = (labelWidthMm / MM_PER_INCH) * DPI;

  const canvas = await renderComponentToCanvas(label, labelWidthPx);
  const imgData = canvas.toDataURL('image/png');

  const canvasWidthMm = PX_TO_MM(canvas.width / CANVAS_SCALE);
  const canvasHeightMm = PX_TO_MM(canvas.height / CANVAS_SCALE);

  // Create an A4 document
  const doc = new jsPDF({ unit: 'mm', format: 'a4' });
  const pageWidth = doc.internal.pageSize.getWidth();
  
  // Center the label on the A4 page
  const x = (pageWidth - canvasWidthMm) / 2;
  const y = 10; // Margin from top

  doc.addImage(imgData, 'PNG', x, y, canvasWidthMm, canvasHeightMm);
  return doc.output('blob');
}

export async function generateLabelsPdf(labels: LabelData[]): Promise<Blob> {
  if (!labels.length) throw new Error('Tidak ada label untuk dicetak.');

  const doc = new jsPDF({ unit: 'mm', format: 'a4' });
  const pageW = doc.internal.pageSize.getWidth();
  const pageH = doc.internal.pageSize.getHeight();
  const margin = 8;
  const cols = 2;
  const rows = 2;
  const labelsPerPage = cols * rows;

  const labelWidthMm = (pageW - margin * (cols + 1)) / cols;
  const labelWidthPx = (labelWidthMm / MM_PER_INCH) * DPI;

  for (let i = 0; i < labels.length; i++) {
    if (i > 0 && i % labelsPerPage === 0) {
      doc.addPage();
    }

    const label = labels[i];
    const canvas = await renderComponentToCanvas(label, labelWidthPx);
    const imgData = canvas.toDataURL('image/png');
    
    const canvasHeightMm = (labelWidthMm / (canvas.width / CANVAS_SCALE)) * (canvas.height / CANVAS_SCALE);

    const localIdx = i % labelsPerPage;
    const c = localIdx % cols;
    const r = Math.floor(localIdx / cols);

    const x = margin + c * (labelWidthMm + margin);
    const y = margin + r * (canvasHeightMm + margin);

    if (y + canvasHeightMm > pageH) { // Basic overflow check
      doc.addPage();
      // Reset position for new page (this label will be the first)
      const newX = margin;
      const newY = margin;
      doc.addImage(imgData, 'PNG', newX, newY, labelWidthMm, canvasHeightMm);
    } else {
      doc.addImage(imgData, 'PNG', x, y, labelWidthMm, canvasHeightMm);
    }
  }

  return doc.output('blob');
}