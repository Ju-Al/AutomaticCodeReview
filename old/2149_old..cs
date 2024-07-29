        public void PngCoreOctree()
        {
            using var memoryStream = new MemoryStream();
            var options = new PngEncoder { Quantizer = KnownQuantizers.Octree };
            this.bmpCore.SaveAsPng(memoryStream, options);
        }

        [Benchmark(Description = "ImageSharp Octree NoDither Png")]
        public void PngCoreOctreeNoDither()
        {
            using var memoryStream = new MemoryStream();
            var options = new PngEncoder { Quantizer = new OctreeQuantizer(new QuantizerOptions { Dither = null }) };
            this.bmpCore.SaveAsPng(memoryStream, options);
        }

        [Benchmark(Description = "ImageSharp Palette Png")]
        public void PngCorePalette()
        {
            using var memoryStream = new MemoryStream();
            var options = new PngEncoder { Quantizer = KnownQuantizers.WebSafe };
            this.bmpCore.SaveAsPng(memoryStream, options);
        }

        [Benchmark(Description = "ImageSharp Palette NoDither Png")]
        public void PngCorePaletteNoDither()
        {
            using var memoryStream = new MemoryStream();
            var options = new PngEncoder { Quantizer = new WebSafePaletteQuantizer(new QuantizerOptions { Dither = null }) };
            this.bmpCore.SaveAsPng(memoryStream, options);
        }

        [Benchmark(Description = "ImageSharp Wu Png")]
        public void PngCoreWu()
        {
            using var memoryStream = new MemoryStream();
            var options = new PngEncoder { Quantizer = KnownQuantizers.Wu };
            this.bmpCore.SaveAsPng(memoryStream, options);
        }

        [Benchmark(Description = "ImageSharp Wu NoDither Png")]
        public void PngCoreWuNoDither()
        {
            using var memoryStream = new MemoryStream();
            var options = new PngEncoder { Quantizer = new WuQuantizer(new QuantizerOptions { Dither = null }), ColorType = PngColorType.Palette };
            this.bmpCore.SaveAsPng(memoryStream, options);
        }
    }
}
