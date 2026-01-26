#!/usr/bin/env python3
"""
Fix discussion prompts and JavaScript quote escaping in English chapter files.
"""

import os
import re

# Mapping of Spanish discussion prompts to English translations
DISCUSSION_TRANSLATIONS = {
    # Chapter 1
    "Compartan juntos sus conclusiones y busquen ejemplos en su relación donde se vean sus diferencias.":
        "Share thoughts of your relationship with each other and provide examples of your most distinct perceptions.",

    # Chapter 2
    "¿Cómo define cada una la felicidad?":
        "How do both of you, individually, define happiness?",
    "Si cada uno se siente feliz. ¿Cómo está el nivel de autoestima de cada uno y cuál es la conexión entre la autoestima y la felicidad?":
        "Do both of you feel happy? How would you describe your self-esteem, and how does it relate to happiness?",

    # Chapter 3
    "Para el hombre, ¿Cómo percibes las P en tu vida?":
        "Husband - How do you perceive the \"P's\" (Power, Prestige, Pleasure) in your life?",
    "Para la mujer, ¿Cómo percibes que tu esposo satisface las P que necesita?":
        "Wife - How do you see your husband looking for his P's?",

    # Chapter 4
    "Para la mujer, ¿Cómo demuestras la importancia de las A en tu vida?":
        "How does the wife demonstrate the importance of the \"A's\" (Affection, Attention, Appreciation) in her life?",
    "Para el hombre, ¿Cómo percibes las necesidades de tu esposa?":
        "How does the husband understand his wife's needs?",

    # Chapter 5
    "¿Cómo manejarían ustedes esta situación basándose en lo aprendido hasta ahora?":
        "Based on this example, how would you manage a similar situation in your relationship?",

    # Chapter 6
    "Conversen acerca de cómo comprender y valorar las diferentes necesidades de los dos, puede mejorar la convivencia.":
        "Discuss how understanding and valuing each other's different needs can improve your relationship.",

    # Chapter 7
    "¿Qué tipo de actividades fortalecen su relación?":
        "What kinds of activities strengthen your relationship as a couple?",

    # Chapter 8
    "Conversen acerca de como cada uno de ustedes demuestra su amor al otro y cómo pueden mejorar la situación.":
        "Discuss how each of you shows love and how your relationship could be improved.",
    "¿Cómo le demuestras tu amor?":
        "How do you show your love?",
    "¿Cómo pueden mejorar la situación?":
        "How could your relationship be improved?",

    # Chapter 9
    "¿Cuál es tu lenguaje de amor en las diferentes épocas del matrimonio?":
        "What has been your main \"love language\" over the years?",
    "¿Cuál ha sido tu lenguaje de amor principal a lo largo de los años?":
        "What has been your main \"love language\" over the years?",

    # Chapter 10
    "Compartan sus respuestas.":
        "Share your responses with each other.",
    "Comparta sus respuestas.":
        "Share your responses with each other.",

    # Chapter 11
    "Identifiquen dos cosas que les hagan más felices en la vida y que no dependan de alguien más.":
        "Identify two things that make you happier in life that don't depend on someone else.",

    # Chapter 12
    "Conversen acerca de una ocasión donde entendieron las necesidades del otro":
        "Discuss a time when you understood each other's needs.",
    "Conversen acerca de una ocasión donde entendieron las necesidades del otro.":
        "Discuss a time when you understood each other's needs.",

    # Chapter 13
    "Menciona dos cosas que puedes cambiar en tu relación después de ver este video.":
        "Two things you can change in your relationship after watching this video.",

    # Chapter 14
    "Conversen:":
        "Discuss:",
    "¿Cómo manejamos las \"deficiencias\" del otro?":
        "How do you handle each other's \"flaws\"?",

    # Chapter 15
    "¿Que diferencias culturales tienen?":
        "What cultural differences do you have?",
    "¿Cómo afecta esto en tu relación?":
        "How does this affect your relationship?",
    "¿Cómo pueden manejarlas?":
        "How can you manage them?",

    # Chapter 16
    "Conversen acerca de como pueden reducir este patrón.":
        "Discuss what you could both do to reduce that pattern.",

    # Chapter 17
    "¿Cómo valoran la opinión del otro?":
        "How do you value each other's opinion?",

    # Chapter 18
    "Compartan los resultados y vean cómo pueden mejorar la comunicación.":
        "Share the results and see how you can improve communication.",

    # Chapter 19
    "Compartan y conversen acerca de cómo pueden mejorar el manejo de sus conflictos.":
        "Share and discuss how you can improve your conflict management.",

    # Chapter 20
    "Conversen acerca de cómo pueden mejorar su capacidad de perdonar.":
        "Discuss how you can improve your ability to forgive.",

    # Chapter 21
    "Conversen acerca de qué pueden hacer para mantener una relación sana y feliz.":
        "Discuss what you can do to maintain a healthy and happy relationship.",

    # Chapter 22
    "Conversen acerca de cómo se sienten cada uno en relación con la intimidad en el matrimonio y qué pueden hacer si es necesario para mejorarla.":
        "Discuss how each one feels regarding the intimacy in the marriage and what can be done if necessary to improve it.",
}


def fix_chapter_file(filepath):
    """Fix discussion prompts and JavaScript quotes in a chapter file."""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    original_content = content

    # Replace Spanish discussion prompts with English
    for spanish, english in DISCUSSION_TRANSLATIONS.items():
        content = content.replace(spanish, english)

    # Fix JavaScript quote escaping
    # Find surveyQuestions block and escape quotes properly
    def escape_js_quotes(match):
        full_match = match.group(0)
        # Replace unescaped " inside the string value (not the delimiters)
        key = match.group(1)
        value = match.group(2)
        # Escape double quotes inside the value
        escaped_value = value.replace('"', '\\"')
        return f'{key}: "{escaped_value}"'

    # Pattern to match q1: "...", q2: "...", etc.
    # But we need a different approach - find the whole surveyQuestions block
    # and process it carefully

    # Find all instances of string values in surveyQuestions that have unescaped quotes
    pattern = r'(q\d+|conversation\d*):\s*"([^"]*)"'

    def process_match(m):
        key = m.group(1)
        value = m.group(2)
        # This value should already have internal quotes escaped
        return f'{key}: "{value}"'

    # Actually, the issue is simpler - quotes inside the value aren't being escaped
    # Let's look for patterns like:  q3: "Q3: Is "happiness" attainable
    # and replace them with:         q3: "Q3: Is \"happiness\" attainable

    # Fix patterns where there are quotes inside JS string literals
    lines = content.split('\n')
    fixed_lines = []
    in_survey_questions = False

    for line in lines:
        if 'window.surveyQuestions' in line:
            in_survey_questions = True
        elif in_survey_questions and line.strip() == '};':
            in_survey_questions = False

        if in_survey_questions and ':' in line and '"' in line:
            # Check if this line has a key-value pair
            match = re.match(r'^(\s*)(q\d+|conversation\d*):\s*"(.+)"(,?)$', line)
            if match:
                indent = match.group(1)
                key = match.group(2)
                value = match.group(3)
                comma = match.group(4) or ''

                # Count quotes in value - if more than 0, we have internal quotes
                quote_count = value.count('"') - value.count('\\"')
                if quote_count > 0:
                    # Replace unescaped quotes with escaped ones
                    # But be careful not to double-escape
                    value = re.sub(r'(?<!\\)"', '\\"', value)

                line = f'{indent}{key}: "{value}"{comma}'

        fixed_lines.append(line)

    content = '\n'.join(fixed_lines)

    if content != original_content:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        return True
    return False


def main():
    for chapter_num in range(1, 23):
        filepath = f'chapter_{chapter_num:02d}.html'
        if os.path.exists(filepath):
            if fix_chapter_file(filepath):
                print(f"Fixed {filepath}")
            else:
                print(f"No changes needed for {filepath}")


if __name__ == '__main__':
    main()
