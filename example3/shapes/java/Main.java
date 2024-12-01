package example3.shapes.java;

import java.util.ArrayList;
import java.util.List;

public class Main {
  public static void main(String[] args) {
    List<Shape> shapes = new ArrayList<>();
    shapes.add(new Circle(2));
    shapes.add(new Square(5));
    shapes.add(new Square(6));

    AreaCalculator calculator = new AreaCalculator(shapes);
    SumCalculator sumCalculator = new SumCalculator(calculator);
    System.out.println(sumCalculator.toHTML());

    List<Shape> solidShapes = new ArrayList<>();
    solidShapes.add(new Circle(3));
    VolumeCalculator volumeCalculator = new VolumeCalculator(solidShapes);

    SumCalculator volumeOutputter = new SumCalculator(volumeCalculator);
    System.out.println(volumeOutputter.toHTML());
  }
}