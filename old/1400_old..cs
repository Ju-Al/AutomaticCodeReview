        {
            MappedDiagnosticsContext.Clear();

            SimpleLayout l = "${message:wrapline=3}";

            var le = LogEventInfo.Create(LogLevel.Info, "logger", "foobar");

            Assert.Equal("foo" + System.Environment.NewLine + "bar", l.Render(le));
        }

        [Fact]
        public void WrapLineMultipleTimesTest()
        {
            MappedDiagnosticsContext.Clear();

            SimpleLayout l = "${message:wrapline=3}";

            var le = LogEventInfo.Create(LogLevel.Info, "logger", "foobarbaz");

            Assert.Equal("foo" + System.Environment.NewLine + "bar" + System.Environment.NewLine + "baz", l.Render(le));
        }

        [Fact]
        public void WrapLineAtPositionAtExactTextLengthTest()
        {
            MappedDiagnosticsContext.Clear();

            SimpleLayout l = "${message:wrapline=6}";

            var le = LogEventInfo.Create(LogLevel.Info, "logger", "foobar");

            Assert.Equal("foobar", l.Render(le));
        }

        [Fact]
        public void WrapLineAtPositionSmallerThanTextLengthTest()
        {
            MappedDiagnosticsContext.Clear();

            SimpleLayout l = "${message:wrapline=7}";

            var le = LogEventInfo.Create(LogLevel.Info, "logger", "foobar");

            Assert.Equal("foobar", l.Render(le));
        }

        [Fact]
        public void WrapLineFromConfig()
        {
            MappedDiagnosticsContext.Clear();

            var configuration = CreateConfigurationFromString(@"
<nlog throwExceptions='true'>
    <targets>
        <target name='d1' type='Debug' layout='${message:wrapline=3}' />
    </targets>
    <rules>
      <logger name=""*"" minlevel=""Trace"" writeTo=""d1"" />
    </rules>
</nlog>");

            var d1 = configuration.FindTargetByName("d1") as DebugTarget;
            Assert.NotNull(d1);
            var layout = d1.Layout as SimpleLayout;
            Assert.NotNull(layout);

            var result = layout.Render(new LogEventInfo(LogLevel.Info, "Test", "foobar"));
            Assert.Equal("foo" + System.Environment.NewLine + "bar", result);
        }
    }
}