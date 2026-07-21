import re
import sys

def main():
    try:
        with open('static/css/style.css', 'r') as f:
            css = f.read()
    except FileNotFoundError:
        print("style.css not found, skipping UI-KIT injection")
        return

    try:
        with open('docs/UI-KIT.html', 'r') as f:
            html = f.read()
    except FileNotFoundError:
        print("UI-KIT.html not found, skipping injection")
        return

    # Look for the marker block
    pattern = r'<!-- KUMO_CSS_START -->.*?<!-- KUMO_CSS_END -->'
    replacement = f'<!-- KUMO_CSS_START -->\n  <style>\n{css}\n  </style>\n  <!-- KUMO_CSS_END -->'

    if re.search(pattern, html, flags=re.DOTALL):
        new_html = re.sub(pattern, replacement, html, flags=re.DOTALL)
        with open('docs/UI-KIT.html', 'w') as f:
            f.write(new_html)
        print("Successfully injected latest style.css into UI-KIT.html")
    else:
        print("Could not find KUMO_CSS markers in UI-KIT.html")

if __name__ == '__main__':
    main()
