package example2.shapes;

public class OutputFormatter {
    private AreaCalculator calculator;

    public OutputFormatter(AreaCalculator calculator) {
        this.calculator = calculator;
    }

    public String toJSON() {
        return String.format("{\"sum\": %.2f}", calculator.sum());
    }

    public String toHTML() {
        return "<p>Sum of the areas of provided shapes: " + calculator.sum() + "</p>";
    }
}