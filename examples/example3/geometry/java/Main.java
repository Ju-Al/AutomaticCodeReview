package example3.geometry.java;

public class Main {
    public static void main(String[] args) {
        Point p1 = new Point(0, 0);
        Point p2 = new Point(4, 4);
        Point testPoint = new Point(2, 2);

        Line line = new Line(p1, p2);
        LineSegment segment = new LineSegment(p1, p2);

        System.out.println("Slope: " + line.getSlope());
        System.out.println("Intercept: " + line.getIntercept());
        System.out.println("Point on line: " + line.isOn(testPoint));
        System.out.println("Segment length: " + segment.getLength());
        System.out.println("Point on segment: " + segment.isOn(testPoint));
    }
}
