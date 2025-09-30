import { LabelData } from '../types';

declare global {
  interface Window {
    jspdf: any;
    html2canvas: any;
  }
}

export const generatePdf = async (labels: LabelData[]): Promise<void> => {
  if (!labels || labels.length === 0) {
    alert("No labels to print.");
    return;
  }

  const { jsPDF } = window.jspdf;
  let pdf: any = null; // Will be initialized on the first label

  // Constants for conversion
  const MM_PER_INCH = 25.4;
  const DPI = 96; // Standard screen DPI assumption
  const PIXELS_TO_MM = (pixels: number) => (pixels / DPI) * MM_PER_INCH;
  const CANVAS_SCALE = 3; // Must match the scale used in html2canvas

  for (let i = 0; i < labels.length; i++) {
    const labelId = `label-preview-${labels[i].id}`;
    const element = document.getElementById(labelId);

    if (element) {
      const canvas = await window.html2canvas(element, {
        scale: CANVAS_SCALE, // Higher scale for better quality
        useCORS: true,
        logging: false,
      });

      const imgData = canvas.toDataURL('image/png');
      
      // Calculate the actual dimensions in mm, accounting for canvas scale
      const canvasWidthMM = PIXELS_TO_MM(canvas.width / CANVAS_SCALE);
      const canvasHeightMM = PIXELS_TO_MM(canvas.height / CANVAS_SCALE);

      if (i === 0) {
        // For the first page, initialize the PDF with this label's dimensions
        pdf = new jsPDF({
          orientation: 'portrait',
          unit: 'mm',
          format: [canvasWidthMM, canvasHeightMM]
        });
      } else {
        // For subsequent pages, add a new page with the specific dimensions
        if (pdf) {
            pdf.addPage([canvasWidthMM, canvasHeightMM], 'portrait');
        }
      }
      
      // Add the image to fill the entire page
      if (pdf) {
          pdf.addImage(imgData, 'PNG', 0, 0, canvasWidthMM, canvasHeightMM);
      }
    }
  }

  if (pdf) {
    pdf.save('smartseller-labels.pdf');
  } else {
    alert("Could not generate PDF. No valid labels found.");
  }
};