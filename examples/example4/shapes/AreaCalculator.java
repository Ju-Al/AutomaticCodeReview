package example4.shapes;

import java.util.List;

public class AreaCalculator {
    protected List<Shape> shapes;

    public AreaCalculator(List<Shape> shapes) {
        this.shapes = shapes;
    }

    public double sum() {
        double totalArea = 0;
        for (Shape shape : shapes) {
            if (shape instanceof Shape) {
                totalArea += shape.area();
            } else {
                throw new RuntimeException("Invalid Shape: Shape does not implement Shape interface");
            }
        }
        return totalArea;
    }
}