import sys

from PIL import Image


def content_height(image, background, tolerance=6):
    pixels = image.convert("RGB")
    width, height = pixels.size

    for y in range(height - 1, -1, -1):
        row = [pixels.getpixel((x, y)) for x in range(0, width, 4)]
        if any(
            abs(r - background[0]) > tolerance
            or abs(g - background[1]) > tolerance
            or abs(b - background[2]) > tolerance
            for r, g, b in row
        ):
            return y + 1
    return height


def main():
    if len(sys.argv) < 2:
        print("usage: trim.py <image.png> [bottom-padding]", file=sys.stderr)
        return 1

    path = sys.argv[1]
    padding = int(sys.argv[2]) if len(sys.argv) > 2 else 48

    image = Image.open(path)
    background = image.convert("RGB").getpixel((image.width // 2, image.height - 2))

    keep = min(image.height, content_height(image, background) + padding)
    if keep < image.height:
        image.crop((0, 0, image.width, keep)).save(path)

    print(f"{path}: {image.width}x{keep}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
