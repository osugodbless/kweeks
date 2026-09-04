const pptxgen = require('pptxgenjs');
const path = require('path');
const html2pptx = require('/home/godbless/.claude/skills/powerpoint/scripts/html2pptx');

const slidesDir = path.join(__dirname, 'slides');
const outputFile = path.join(__dirname, 'Kweeks-Hackathon-Presentation.pptx');

async function build() {
  const pptx = new pptxgen();
  pptx.layout = 'LAYOUT_16x9';
  pptx.author = 'Godbless';
  pptx.title = 'Kweeks - The Live Money Quiz Platform';
  pptx.subject = 'BMONI Embedded API Hackathon 2026';

  const slideFiles = [
    'slide01-title.html',
    'slide02-problem.html',
    'slide03-solution.html',
    'slide04-howitworks.html',
    'slide05-market.html',
    'slide06-competitors.html',
    'slide07-advantage.html',
    'slide08-bmoni.html',
    'slide09-techstack.html',
    'slide10-demo.html',
    'slide11-vision.html',
    'slide12-thankyou.html',
  ];

  for (const file of slideFiles) {
    const htmlPath = path.join(slidesDir, file);
    console.log('Processing:', file);
    try {
      const { slide, placeholders } = await html2pptx(htmlPath, pptx);
      console.log('  -> OK', placeholders.length, 'placeholders');
    } catch (err) {
      console.error('  -> ERROR:', err.message);
    }
  }

  await pptx.writeFile({ fileName: outputFile });
  console.log('\nPresentation saved to:', outputFile);
}

build().catch(err => {
  console.error('Build failed:', err);
  process.exit(1);
});
