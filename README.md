# x, y, z 2-tuples with volumetric colors in a baseline cube

(x, y), (x, z), (y, z)

* Every axis is one of the RGB. E.g: (x, y) => (r, g)
* The center of the screen is in the origin of the local space of the baseline cube
* For every frame, draw the 3 images

* or just send the origin of the screen

* same camera, scene is an infinite square at a fixed distance from the camera (origin of the local cube space), and colors are how distant from the camera the point is

1. Import/Export HitterList
2. Improve speed by using less jobs (one per line instead of one per pixel)
3. Frames:
    a. 300x300 fully rendered (512 spp, depth = 20)
    b. convert to 300x300 grayscale
    c. 300x300 noisy (4 spp, depth = 20)
    d. convert to 300x300 grayscale
    e. 30x30 camera distance (4 spp, depth = 20)
        Use max distance in the scene

break render into parts:
  lambertian
  specular
  etc
