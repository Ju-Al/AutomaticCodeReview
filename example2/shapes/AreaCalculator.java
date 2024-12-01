package example2.shapes;

import java.util.List;

public class AreaCalculator {
    private List<Object> shapes;

    public AreaCalculator(List<Object> shapes) {
        this.shapes = shapes;
    }

    public double sum() {
        double totalArea = 0;

        for (Object shape : shapes) {
            if (shape instanceof Square) {
                Square square = (Square) shape;
                totalArea += Math.pow(square.getLength(), 2);
            } else if (shape instanceof Circle) {
                Circle circle = (Circle) shape;
                totalArea += Math.PI * Math.pow(circle.getRadius(), 2);
            }
        }

        return totalArea;
    }
}