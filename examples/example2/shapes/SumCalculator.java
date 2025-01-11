package example2.shapes;

/* import org.json.JSONObject; */

public class SumCalculator {
    private AreaCalculator calculator;

    public SumCalculator(AreaCalculator calculator) {
        this.calculator = calculator;
    }

    /*
     * public String toJSON() {
     * JSObject data = new JSObject();
     * data.put("sum", calculator.sum());
     * return data.toString();
     * }
     */

    // Method to return HTML representation
    public String toHTML() {
        return "<p>Sum of the areas of provided shapes: " + calculator.sum() + "</p>";
    }
}