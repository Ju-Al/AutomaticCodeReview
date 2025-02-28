package example2.shapes;

import java.util.ArrayList;
import java.util.List;

public class Main {
  public static void main(String[] args) {
    List<Object> shapes = new ArrayList<>();
    shapes.add(new Circle(2));
    shapes.add(new Square(5));
    shapes.add(new Square(6));

    AreaCalculator calculator = new AreaCalculator(shapes);
    OutputFormatter sumCalculator = new OutputFormatter(calculator);
    System.out.println(sumCalculator.toHTML());
  }
}