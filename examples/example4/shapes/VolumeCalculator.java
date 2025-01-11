package example4.shapes;

import java.util.List;

public class VolumeCalculator extends AreaCalculator {
    public VolumeCalculator(List<Shape> shapes) {
        super(shapes);
    }

    @Override
    public double sum() {
        double totalVolume = 0;
        for (Shape shape : shapes) {
            totalVolume += shape.volume();
        }
        return totalVolume;
    }
}
