package example2.shapes;

public class SumCalculator {
    private AreaCalculator calculator;

    public SumCalculator(AreaCalculator calculator) {
        this.calculator = calculator;
    }

    public String toJSON() {
        return String.format("{\"sum\": %.2f}", calculator.sum());
    }

    public String toHTML() {
        return "<p>Sum of the areas of provided shapes: " + calculator.sum() + "</p>";
    }
}