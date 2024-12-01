package example3.shapes.java;

import java.util.List;

public class VolumeCalculator extends AreaCalculator {
    public VolumeCalculator(List<Shape> shapes) {
        super(shapes);
    }

    @Override
    public double sum() {
        double totalVolume = 0;
        for (Shape shape : shapes) {
            totalVolume += shape.area();
        }
        return totalVolume;
    }
}
