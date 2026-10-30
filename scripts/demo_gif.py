#!/usr/bin/env python3
"""Transforma os quadros ANSI de TestDemoFrames em docs/demo.gif.

Uso: python3 scripts/demo_gif.py <pasta-dos-quadros> <saida.gif>
Requer Pillow e as fontes DejaVu (pacote fonts-dejavu no Debian/Ubuntu).
Normalmente é chamado por `make demo-gif`.
"""
import glob
import os
import re
import sys
import unicodedata

from PIL import Image, ImageDraw, ImageFont

FONT_DIR = "/usr/share/fonts/truetype/dejavu"
FONT_SIZE = 16
LINE_HEIGHT = 21
MARGIN = 24
TITLE_BAR = 34

BG = (30, 30, 46)
FG = (205, 214, 244)
WINDOW_BORDER = (69, 71, 90)
TRAFFIC = [(243, 139, 168), (249, 226, 175), (166, 227, 161)]

mono = ImageFont.truetype(f"{FONT_DIR}/DejaVuSansMono.ttf", FONT_SIZE)
mono_bold = ImageFont.truetype(f"{FONT_DIR}/DejaVuSansMono-Bold.ttf", FONT_SIZE)
# A DejaVu Sans Mono não tem braille (o spinner); a DejaVu Sans tem.
fallback = ImageFont.truetype(f"{FONT_DIR}/DejaVuSans.ttf", FONT_SIZE)

CELL_W = round(mono.getlength("M"))
SGR = re.compile(r"\x1b\[([0-9;]*)m")


def parse(text):
    """Devolve linhas de células (char, cor, fundo, negrito, riscado)."""
    lines = []
    for raw in text.split("\n"):
        cells, pos = [], 0
        fg, bg, bold, strike = None, None, False, False
        for m in SGR.finditer(raw + "\x1b[m"):
            for ch in raw[pos:m.start()]:
                cells.append((ch, fg, bg, bold, strike))
                if unicodedata.east_asian_width(ch) in "WF":
                    cells.append(("", fg, bg, bold, strike))
            pos = m.end()
            params = [int(p) if p else 0 for p in m.group(1).split(";")]
            i = 0
            while i < len(params):
                p = params[i]
                if p == 0:
                    fg, bg, bold, strike = None, None, False, False
                elif p == 1:
                    bold = True
                elif p == 22:
                    bold = False
                elif p == 9:
                    strike = True
                elif p == 29:
                    strike = False
                elif p == 39:
                    fg = None
                elif p == 49:
                    bg = None
                elif p == 38 and i + 4 < len(params) and params[i + 1] == 2:
                    fg = tuple(params[i + 2:i + 5])
                    i += 4
                elif p == 48 and i + 4 < len(params) and params[i + 1] == 2:
                    bg = tuple(params[i + 2:i + 5])
                    i += 4
                i += 1
        lines.append(cells)
    return lines


# Traços de cada caractere de borda: (cima, baixo, esquerda, direita), com
# peso 1 (fino), 3 (grosso) ou "d" (duplo).
SEGMENTS = {
    "─": (0, 0, 1, 1), "│": (1, 1, 0, 0),
    "┌": (0, 1, 0, 1), "┐": (0, 1, 1, 0), "└": (1, 0, 0, 1), "┘": (1, 0, 1, 0),
    "━": (0, 0, 3, 3), "┃": (3, 3, 0, 0),
    "┏": (0, 3, 0, 3), "┓": (0, 3, 3, 0), "┗": (3, 0, 0, 3), "┛": (3, 0, 3, 0),
    "┣": (3, 3, 0, 3), "┫": (3, 3, 3, 0),
    "╻": (0, 3, 0, 0), "╹": (3, 0, 0, 0), "╺": (0, 0, 0, 3), "╸": (0, 0, 3, 0),
    "═": (0, 0, "d", "d"), "║": ("d", "d", 0, 0),
    "╔": (0, "d", 0, "d"), "╗": (0, "d", "d", 0), "╚": ("d", 0, 0, "d"), "╝": ("d", 0, "d", 0),
}


def blend(color, amount):
    return tuple(round(c * amount + b * (1 - amount)) for c, b in zip(color, BG))


def draw_box_char(d, ch, x, y, color):
    """Desenha blocos e bordas como formas para não ficar fresta entre linhas."""
    w, h = CELL_W, LINE_HEIGHT
    cx, cy = x + w // 2, y + h // 2
    r = 5
    if ch == "█":
        d.rectangle([x, y, x + w - 1, y + h - 1], fill=color)
    elif ch == "▓":
        d.rectangle([x, y, x + w - 1, y + h - 1], fill=blend(color, 0.7))
    elif ch in "╔╗╚╝":
        dx = 1 if ch in "╔╚" else -1
        dy = 1 if ch in "╔╗" else -1
        x_edge = x + w if dx > 0 else x
        y_edge = y + h if dy > 0 else y
        for off in (-2, 2):
            vx, hy = cx + off * dx, cy + off * dy
            d.line([vx, hy, vx, y_edge], fill=color)
            d.line([vx, hy, x_edge, hy], fill=color)
    elif ch in SEGMENTS:
        up, down, left, right = SEGMENTS[ch]
        for weight, (x0, y0, x1, y1) in (
            (up, (cx, y, cx, cy)), (down, (cx, cy, cx, y + h)),
            (left, (x, cy, cx, cy)), (right, (cx, cy, x + w, cy)),
        ):
            vertical = x0 == x1
            if weight == 1:
                d.line([x0, y0, x1, y1], fill=color)
            elif weight == 3:
                if vertical:
                    d.rectangle([x0 - 1, y0, x0 + 1, y1], fill=color)
                else:
                    d.rectangle([x0, y0 - 1, x1, y0 + 1], fill=color)
            elif weight == "d":
                for off in (-2, 2):
                    if vertical:
                        d.line([x0 + off, y0, x1 + off, y1], fill=color)
                    else:
                        d.line([x0, y0 + off, x1, y1 + off], fill=color)
    elif ch == "╭":
        d.line([cx + r, cy, x + w, cy], fill=color)
        d.line([cx, cy + r, cx, y + h], fill=color)
        d.arc([cx, cy, cx + 2 * r, cy + 2 * r], 180, 270, fill=color)
    elif ch == "╮":
        d.line([x, cy, cx - r, cy], fill=color)
        d.line([cx, cy + r, cx, y + h], fill=color)
        d.arc([cx - 2 * r, cy, cx, cy + 2 * r], 270, 360, fill=color)
    elif ch == "╰":
        d.line([cx, y, cx, cy - r], fill=color)
        d.line([cx + r, cy, x + w, cy], fill=color)
        d.arc([cx, cy - 2 * r, cx + 2 * r, cy], 90, 180, fill=color)
    elif ch == "╯":
        d.line([cx, y, cx, cy - r], fill=color)
        d.line([x, cy, cx - r, cy], fill=color)
        d.arc([cx - 2 * r, cy - 2 * r, cx, cy], 0, 90, fill=color)
    else:
        return False
    return True


def render(lines, cols, rows):
    width = MARGIN * 2 + cols * CELL_W
    height = TITLE_BAR + MARGIN + rows * LINE_HEIGHT
    img = Image.new("RGB", (width, height), BG)
    d = ImageDraw.Draw(img)

    # Moldura de janela de terminal.
    d.rectangle([0, 0, width - 1, height - 1], outline=WINDOW_BORDER)
    for i, c in enumerate(TRAFFIC):
        x = 18 + i * 20
        d.ellipse([x, 12, x + 11, 23], fill=c)
    title = "painel"
    d.text(((width - mono.getlength(title)) / 2, 9), title, font=mono, fill=(127, 132, 156))

    for row, cells in enumerate(lines[:rows]):
        y = TITLE_BAR + row * LINE_HEIGHT
        for col, (ch, fg, bg, bold, strike) in enumerate(cells):
            x = MARGIN + col * CELL_W
            if bg:
                d.rectangle([x, y, x + CELL_W - 1, y + LINE_HEIGHT - 1], fill=bg)
            if not ch or ch == " ":
                continue
            color = fg or FG
            if not draw_box_char(d, ch, x, y, color):
                font = fallback if 0x2800 <= ord(ch) <= 0x28FF else (mono_bold if bold else mono)
                d.text((x, y + 2), ch, font=font, fill=color)
            if strike:
                d.line([x, y + LINE_HEIGHT // 2, x + CELL_W, y + LINE_HEIGHT // 2], fill=color)
    return img


def main(frames_dir, out):
    paths = sorted(glob.glob(os.path.join(frames_dir, "frame_*.ans")))
    if not paths:
        sys.exit(f"nenhum quadro em {frames_dir}")
    parsed = [parse(open(p, encoding="utf-8").read()) for p in paths]

    # Remove linhas em branco no fim, comuns a todos os quadros.
    rows = max(max((i + 1 for i, l in enumerate(f) if any(c[0].strip() for c in l)), default=0) for f in parsed) + 1
    cols = max(len(l) for f in parsed for l in f)
    frames = [render(f, cols, rows) for f in parsed]

    # Uma paleta só para todos os quadros, para as cores não oscilarem.
    strip = Image.new("RGB", (frames[0].width, frames[0].height * len(frames)))
    for i, f in enumerate(frames):
        strip.paste(f, (0, i * f.height))
    # Pillow < 9.1 não tem os enums Quantize/Dither.
    quantize = getattr(Image, "Quantize", Image)
    dither = getattr(Image, "Dither", Image)
    palette = strip.quantize(colors=255, method=quantize.MEDIANCUT)
    gif = [f.quantize(palette=palette, dither=dither.NONE) for f in frames]

    durations = [1000] * len(gif)
    durations[-1] = 3000  # segura o último quadro antes de recomeçar
    gif[0].save(out, save_all=True, append_images=gif[1:], duration=durations, loop=0, optimize=True)
    print(f"{out}: {len(gif)} quadros, {frames[0].width}x{frames[0].height}, {os.path.getsize(out) // 1024} KB")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    main(sys.argv[1], sys.argv[2])
