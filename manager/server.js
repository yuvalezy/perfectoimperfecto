const express = require('express');
const path = require('path');
const fs = require('fs');
const { execSync } = require('child_process');
const cheerio = require('cheerio');

const app = express();
const PORT = 3456;
const ROOT_DIR = path.resolve(__dirname, '..');

app.use(express.json({ limit: '10mb' }));
app.use(express.static(path.join(__dirname, 'public')));

// ---------------------------------------------------------------------------
// Helpers – READ (uses cheerio for parsing only)
// ---------------------------------------------------------------------------

function getChapterFiles() {
  const files = fs.readdirSync(ROOT_DIR).filter(f => /^(capitulo|chapter)_\d+\.html$/.test(f));
  files.sort((a, b) => {
    const numA = parseInt(a.match(/\d+/)[0], 10);
    const numB = parseInt(b.match(/\d+/)[0], 10);
    const langA = a.startsWith('capitulo') ? 0 : 1;
    const langB = b.startsWith('capitulo') ? 0 : 1;
    if (langA !== langB) return langA - langB;
    return numA - numB;
  });
  return files;
}

function parseChapter(filename) {
  const filePath = path.join(ROOT_DIR, filename);
  const html = fs.readFileSync(filePath, 'utf-8');
  const $ = cheerio.load(html, { decodeEntities: false });

  const lang = $('html').attr('lang') || (filename.startsWith('capitulo') ? 'es' : 'en');
  const title = $('title').text();
  const mainInstruction = $('h1').first().text();

  const questions = [];
  $('form#pollForm > .question-section').each((i, section) => {
    const $section = $(section);
    const questionTitle = $section.find('.question-title').first().text().trim();

    if ($section.find('input[type="email"]').length > 0) return;
    if ($section.find('.g-recaptcha').length > 0) return;

    const radios = $section.find('input[type="radio"]');
    const checkboxes = $section.find('input[type="checkbox"]');
    const textarea = $section.find('textarea');

    if (radios.length > 0) {
      const name = radios.first().attr('name');
      const options = [];
      radios.each((j, radio) => options.push($(radio).attr('value')));
      questions.push({
        type: 'radio',
        name,
        title: questionTitle,
        options,
        conditional: $section.hasClass('conditional-section'),
        sectionId: $section.attr('id') || ''
      });
    } else if (checkboxes.length > 0) {
      const name = checkboxes.first().attr('name');
      const options = [];
      checkboxes.each((j, cb) => options.push($(cb).attr('value')));
      questions.push({
        type: 'checkbox',
        name,
        title: questionTitle,
        options,
        conditional: $section.hasClass('conditional-section'),
        sectionId: $section.attr('id') || ''
      });
    } else if (textarea.length > 0) {
      const name = textarea.attr('name') || 'conversation';
      questions.push({
        type: 'textarea',
        name,
        title: questionTitle,
        label: $section.find('.text-input-section label').text().trim(),
        placeholder: textarea.attr('placeholder') || ''
      });
    }
  });

  // Extract email settings
  const emailLabel = $('form#pollForm label[for="email"]').text().trim() || 'Email';
  const emailPlaceholder = $('form#pollForm input[type="email"]').attr('placeholder') || '';
  const submitText = $('form#pollForm .submit-btn').text().trim() || 'Submit';
  const resetText = $('form#pollForm .reset-btn').text().trim() || 'Clear';
  const successMsg = $('.success-message').text().trim() || '';
  const summaryTitle = $('.summary-title').text().trim() || '';

  const chapterName = extractScriptVar(html, 'chapterName') || title;

  // Detect if file has inline checkbox/conditional styles
  const hasCheckboxStyles = html.includes('.checkbox-group') || questions.some(q => q.type === 'checkbox');
  // Detect if file has conditional logic script
  const conditionalScript = extractConditionalScript(html);

  return {
    filename, lang, title, chapterName, mainInstruction, questions,
    emailLabel, emailPlaceholder, submitText, resetText,
    successMsg, summaryTitle, hasCheckboxStyles, conditionalScript
  };
}

function extractScriptVar(html, varName) {
  const regex = new RegExp(`window\\.${varName}\\s*=\\s*["']([^"']+)["']`);
  const match = html.match(regex);
  return match ? match[1] : null;
}

function extractConditionalScript(html) {
  // Extract the conditional logic block (e.g. Q1->Q2 show/hide)
  const match = html.match(/\/\/ Handle conditional questions\s*\n\s*document\.addEventListener\('DOMContentLoaded',\s*function\(\)\s*\{([\s\S]*?)\}\);/);
  if (match) return match[0];
  return '';
}

// ---------------------------------------------------------------------------
// Helpers – WRITE (generates HTML from scratch via templates)
// ---------------------------------------------------------------------------

function esc(str) {
  if (!str) return '';
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function escAttr(str) {
  if (!str) return '';
  return str.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function generateChapterHtml(data) {
  const lang = data.lang || 'es';
  const hasCheckboxes = data.questions.some(q => q.type === 'checkbox');
  const hasConditionals = data.questions.some(q => q.conditional);

  // Build inline styles if needed
  let inlineStyles = '';
  if (hasCheckboxes || hasConditionals) {
    inlineStyles = `
    <style>
        .checkbox-group {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        .checkbox-option {
            display: flex;
            align-items: center;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.3s ease;
        }

        .checkbox-option:hover {
            border-color: #667eea;
            background-color: #f5f7ff;
        }

        .checkbox-option input[type="checkbox"] {
            margin-right: 12px;
            cursor: pointer;
            width: 18px;
            height: 18px;
        }

        .checkbox-option label {
            cursor: pointer;
            flex: 1;
            color: #333;
        }

        .checkbox-option input[type="checkbox"]:checked + label {
            color: #667eea;
            font-weight: 500;
        }

        .hidden {
            display: none;
        }

        .conditional-section {
            transition: opacity 0.3s ease;
        }
    </style>`;
  }

  // Build question sections
  let questionsHtml = '';
  data.questions.forEach((q, i) => {
    if (q.type === 'radio') {
      const condClass = q.conditional ? ' conditional-section hidden' : '';
      const idAttr = q.sectionId ? ` id="${q.sectionId}"` : '';
      let optsHtml = '';
      q.options.forEach((opt, j) => {
        const optId = `${q.name}_${j + 1}`;
        const req = j === 0 && !q.conditional ? ' required' : '';
        optsHtml += `
                    <div class="option">
                        <input type="radio" id="${optId}" name="${q.name}" value="${escAttr(opt)}"${req}>
                        <label for="${optId}">${esc(opt)}</label>
                    </div>`;
      });
      questionsHtml += `
            <!-- ${q.name.toUpperCase()} -->
            <div class="question-section${condClass}"${idAttr}>
                <div class="question-title">${esc(q.title)}</div>
                <div class="options">${optsHtml}
                </div>
            </div>
`;
    } else if (q.type === 'checkbox') {
      const condClass = q.conditional ? ' conditional-section hidden' : '';
      const idAttr = q.sectionId ? ` id="${q.sectionId}"` : '';
      let optsHtml = '';
      q.options.forEach((opt, j) => {
        const optId = `${q.name}_${j + 1}`;
        optsHtml += `
                    <div class="checkbox-option">
                        <input type="checkbox" id="${optId}" name="${q.name}" value="${escAttr(opt)}">
                        <label for="${optId}">${esc(opt)}</label>
                    </div>`;
      });
      questionsHtml += `
            <!-- ${q.name.toUpperCase()} -->
            <div class="question-section${condClass}"${idAttr}>
                <div class="question-title">${esc(q.title)}</div>
                <div class="checkbox-group">${optsHtml}
                </div>
            </div>
`;
    } else if (q.type === 'textarea') {
      questionsHtml += `
            <!-- Conversation Section -->
            <div class="question-section">
                <div class="question-title">${esc(q.title)}</div>
                <div class="text-input-section">
                    <label for="${q.name}" style="color: #666; font-size: 14px;">${esc(q.label)}</label>
                    <textarea id="${q.name}" name="${q.name}" placeholder="${escAttr(q.placeholder)}" required></textarea>
                    <div class="error-message" id="textError" style="display: none;"></div>
                </div>
            </div>
`;
    }
  });

  // Build surveyQuestions object
  const sqEntries = data.questions.map(q => {
    const val = q.type === 'textarea' ? q.label : q.title;
    return `            ${q.name}: "${val.replace(/"/g, '\\"')}"`;
  });
  const surveyQuestionsBlock = `{\n${sqEntries.join(',\n')}\n        }`;

  // Build conditional script if needed
  let conditionalScriptBlock = '';
  if (data.conditionalScript && data.conditionalScript.trim()) {
    conditionalScriptBlock = '\n        ' + data.conditionalScript.trim() + '\n';
  }

  const chapterName = data.chapterName || data.title;

  return `<!DOCTYPE html>
<html lang="${lang}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${esc(data.title)}</title>
    <link rel="stylesheet" href="survey.css">${inlineStyles}
</head>
<body>
    <div class="container">
        <h1>${esc(data.mainInstruction)}</h1>
        <form id="pollForm">
${questionsHtml}
            <!-- Email Section -->
            <div class="question-section">
                <div class="question-title"><label for="email">${esc(data.emailLabel)}</label></div>
                <div class="text-input-section">
                    <input type="email" id="email" name="email" placeholder="${escAttr(data.emailPlaceholder)}" required>
                    <div class="error-message" id="emailError" style="display: none;"></div>
                </div>
            </div>

            <!-- reCAPTCHA -->
            <div class="question-section">
                <div class="g-recaptcha" data-sitekey="6LeyoggsAAAAAAgXzEg9PAC9ypZtr-yyc24cAnN_"></div>
                <div class="error-message" id="recaptchaError" style="display: none;"></div>
            </div>

            <!-- Buttons -->
            <div class="button-group">
                <button type="submit" class="submit-btn" id="submitBtn">${esc(data.submitText)}</button>
                <button type="reset" class="reset-btn" id="resetBtn">${esc(data.resetText)}</button>
            </div>
        </form>

        <!-- Success Message -->
        <div class="success-message" id="successMessage">
            ${data.successMsg || (lang === 'es' ? '&#10003; ¡Encuesta enviada correctamente!' : '&#10003; Survey submitted successfully!')}
        </div>

        <!-- Summary Section -->
        <div class="summary-section" id="summarySection" style="display: none;">
            <h2 class="summary-title">${data.summaryTitle || (lang === 'es' ? 'Resumen de tus respuestas' : 'Summary of your answers')}</h2>
            <div id="summaryContent"></div>
        </div>
    </div>

    <script>${conditionalScriptBlock}
        // Define questions for this survey
        window.surveyQuestions = ${surveyQuestionsBlock};

        // Define the chapter name for this survey
        window.chapterName = "${chapterName.replace(/"/g, '\\"')}";
    </script>

    <!-- EmailJS Library -->
    <script src="https://cdn.jsdelivr.net/npm/emailjs-com@3/dist/email.min.js"></script>
    <script>
        (function() {
            emailjs.init("wZ_Z4F9Y-8CcFzD2g");
        })();
    </script>

    <!-- Google reCAPTCHA v2 -->
    <script src="https://www.google.com/recaptcha/api.js" async defer></script>

    <script src="survey.js?v=20250110005"></script>
</body>
</html>
`;
}

// ---------------------------------------------------------------------------
// API Routes
// ---------------------------------------------------------------------------

app.get('/api/chapters', (req, res) => {
  try {
    const files = getChapterFiles();
    const chapters = files.map(f => {
      const data = parseChapter(f);
      return {
        filename: data.filename,
        lang: data.lang,
        title: data.title,
        chapterName: data.chapterName,
        questionCount: data.questions.length
      };
    });
    res.json(chapters);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.get('/api/chapters/:filename', (req, res) => {
  try {
    const filename = req.params.filename;
    if (!/^(capitulo|chapter)_\d+\.html$/.test(filename)) {
      return res.status(400).json({ error: 'Invalid filename' });
    }
    const data = parseChapter(filename);
    res.json(data);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.put('/api/chapters/:filename', (req, res) => {
  try {
    const filename = req.params.filename;
    if (!/^(capitulo|chapter)_\d+\.html$/.test(filename)) {
      return res.status(400).json({ error: 'Invalid filename' });
    }
    const filePath = path.join(ROOT_DIR, filename);
    if (!fs.existsSync(filePath)) {
      return res.status(404).json({ error: 'File not found' });
    }

    const newHtml = generateChapterHtml(req.body);
    fs.writeFileSync(filePath, newHtml, 'utf-8');

    res.json({ success: true, message: `${filename} saved successfully` });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// ---------------------------------------------------------------------------
// Git API
// ---------------------------------------------------------------------------

app.get('/api/git/status', (req, res) => {
  try {
    const status = execSync('git status --porcelain', { cwd: ROOT_DIR, encoding: 'utf-8' });
    const branch = execSync('git branch --show-current', { cwd: ROOT_DIR, encoding: 'utf-8' }).trim();
    res.json({
      branch,
      changes: status.trim().split('\n').filter(l => l.length > 0),
      hasChanges: status.trim().length > 0
    });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.post('/api/git/commit-and-push', (req, res) => {
  try {
    const message = req.body.message || 'Update survey templates';

    const status = execSync('git status --porcelain', { cwd: ROOT_DIR, encoding: 'utf-8' });
    if (!status.trim()) {
      return res.json({ success: true, message: 'No changes to commit' });
    }

    execSync('git add *.html', { cwd: ROOT_DIR, encoding: 'utf-8' });
    execSync(`git commit -m "${message.replace(/"/g, '\\"')}"`, { cwd: ROOT_DIR, encoding: 'utf-8' });

    try {
      const branch = execSync('git branch --show-current', { cwd: ROOT_DIR, encoding: 'utf-8' }).trim();
      execSync(`git push -u origin ${branch}`, { cwd: ROOT_DIR, encoding: 'utf-8', stdio: 'pipe' });
    } catch (pushErr) {
      return res.json({
        success: true,
        message: 'Changes committed locally. Push failed — you may need to configure a remote or check your network.'
      });
    }

    res.json({ success: true, message: 'Changes committed and pushed successfully!' });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.get('/api/git/log', (req, res) => {
  try {
    const log = execSync('git log --oneline -20', { cwd: ROOT_DIR, encoding: 'utf-8' });
    res.json({ log: log.trim().split('\n') });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// ---------------------------------------------------------------------------
// Static files for preview
// ---------------------------------------------------------------------------

app.get('/preview/:filename', (req, res) => {
  const filename = req.params.filename;
  if (!/^(capitulo|chapter)_\d+\.html$/.test(filename)) {
    return res.status(400).send('Invalid filename');
  }
  const filePath = path.join(ROOT_DIR, filename);
  if (fs.existsSync(filePath)) {
    res.sendFile(filePath);
  } else {
    res.status(404).send('File not found');
  }
});

app.get('/survey.css', (req, res) => res.sendFile(path.join(ROOT_DIR, 'survey.css')));
app.get('/survey.js', (req, res) => res.sendFile(path.join(ROOT_DIR, 'survey.js')));

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------
app.listen(PORT, () => {
  console.log('');
  console.log('  ===================================================');
  console.log('  Perfecto Imperfecto - Template Manager');
  console.log('  ===================================================');
  console.log('');
  console.log(`  Open your browser at:  http://localhost:${PORT}`);
  console.log('');
  console.log('  Press Ctrl+C to stop the server.');
  console.log('');
});
