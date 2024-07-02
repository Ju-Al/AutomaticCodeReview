from distutils.core import setup, Extension

setup(ext_modules = cythonize(Extension(
           "rect",                                # the extension name
           sources=["rect.pyx", "Rectangle.cpp"], # the Cython source and
                                                  # additional C++ source files
           language="c++",                        # generate and compile C++ code
      )))
