"""Strip metadata from the documentation images and cap their width.

The framed screenshots come out around 1550px wide but are displayed at a third
of that, so shrinking them keeps the repository small without any visible loss.
Metadata is dropped because it carries nothing useful and everything in these
files should be reproducible from the source screenshots.
"""

import sys

from PIL import Image

MAX_WIDTH = {
    "hero.png": 900,
    "login.png": 780,
    "patch-effort.png": 780,
    "patch-models.png": 780,
    "settings.png": 780,
    "control-panel.png": 900,
}


def optimize(path):
    name = path.split("/")[-1]
    image = Image.open(path)

    limit = MAX_WIDTH.get(name)
    if limit and image.width > limit:
        height = round(image.height * limit / image.width)
        image = image.resize((limit, height), Image.LANCZOS)

    clean = Image.frombytes(image.mode, image.size, image.tobytes())
    clean.save(path, optimize=True)

    return clean.size


def main():
    if len(sys.argv) < 2:
        print("usage: optimize.py <image.png> [...]", file=sys.stderr)
        return 1

    for path in sys.argv[1:]:
        width, height = optimize(path)
        print(f"  {path.split('/')[-1]} ({width}x{height})")
    return 0


if __name__ == "__main__":
    sys.exit(main())
