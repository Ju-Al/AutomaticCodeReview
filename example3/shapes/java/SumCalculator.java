package example3.shapes.java;

/* import org.json.JSONObject; */

public class SumCalculator {
    protected AreaCalculator calculator;

    public SumCalculator(AreaCalculator calculator) {
        this.calculator = calculator;
    }

    public String toJSON() {
        /*
         * JSONObject data = new JSONObject();
         * data.put("sum", calculator.sum());
         * return data.toString();
         */
        return "";
    }

    public String toHTML() {
        return String.format("<p>Sum of the areas of provided shapes: %.2f</p>", calculator.sum());
    }
}