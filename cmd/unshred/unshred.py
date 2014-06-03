#!/usr/bin/python
# A solution to the Instagram Engineering Challenge (http://goo.gl/B92t0) in Python
from PIL import Image
import math

def calc_diff(s0, s1):
    """Calculate difference between right edge of shred s0 and left edge of
    shred s1.""" 
    d = 0.0
    x0 = s0 * shred_width + shred_width - 1
    x1 = s1 * shred_width
    for y in xrange(0, image_height):
        p0 = data[y * image_width + x0]
        p1 = data[y * image_width + x1]
        for channel in xrange(3):
            d += (math.log(p0[channel]/255.0 + 1.0 / 256) -
                  math.log(p1[channel]/255.0 + 1.0 / 256)) ** 2
    return d
     
image = Image.open("TokyoPanoramaShredded.png")
data = image.getdata() 
image_width, image_height = image.size
shred_width = 32
shred_count = image_width / shred_width

edges = sorted((calc_diff(i, j), i, j) for i in xrange(shred_count) for j in xrange(shred_count) if i != j)

join_count = 0
shreds = [[-1, -1] for i in range(shred_count)]
for d, i, j in edges:
    if shreds[i][1] < 0 and shreds[j][0] < 0:
        shreds[i][1] = j
        shreds[j][0] = i
        join_count += 1
        if join_count == shred_count - 1:
            break

i = 0
while shreds[i][0] >= 0:
    i += 1

unshredded = Image.new("RGBA", image.size)
for j in range(shred_count):
    x0 = i * shred_width
    x1 = x0 + shred_width 
    source = image.crop((x0, 0, x1, image_height))
    unshredded.paste(source, (j * shred_width, 0))
    i = shreds[i][1]

unshredded.save("unshredded.jpg", "JPEG")
