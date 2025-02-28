package example4.shapes;

public class Circle implements Shape {
    private double radius;

    public Circle(double radius) {
        this.radius = radius;
    }

    @Override
    public double area() {
        return Math.PI * radius * radius;
    }

    @Override
    public double volume() {
        throw new UnsupportedOperationException("The volume cannot be calculated");
    }
}