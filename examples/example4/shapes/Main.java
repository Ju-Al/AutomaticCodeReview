package example4.shapes;

import java.util.ArrayList;
import java.util.List;

public class Main {
  public static void main(String[] args) {
    List<Shape> shapes = new ArrayList<>();
    shapes.add(new Circle(2));
    shapes.add(new Square(5));
    shapes.add(new Square(6));
    shapes.add(new Dice(2));

    AreaCalculator calculator = new AreaCalculator(shapes);
    SumCalculator sumCalculator = new SumCalculator(calculator);
    System.out.println(sumCalculator.toHTML());

    VolumeCalculator volumeCalculator = new VolumeCalculator(shapes);
    SumCalculator volumeOutputter = new SumCalculator(volumeCalculator);

    System.out.println("Volume HTML Output:");
    System.out.println(volumeOutputter.toHTML());
  }
}