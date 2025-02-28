package example3.geometry.java;

public class LineSegment extends Line {

    public LineSegment(Point p1, Point p2) {
        super(p1, p2);
    }

    public double getLength() {
        double dx = getP2().getX() - getP1().getX();
        double dy = getP2().getY() - getP1().getY();
        return Math.sqrt(dx * dx + dy * dy);
    }

    @Override
    public boolean isOn(Point point) {
        if (!super.isOn(point)) {
            return false;
        }
        double minX = Math.min(getP1().getX(), getP2().getX());
        double maxX = Math.max(getP1().getX(), getP2().getX());
        double minY = Math.min(getP1().getY(), getP2().getY());
        double maxY = Math.max(getP1().getY(), getP2().getY());

        return point.getX() >= minX && point.getX() <= maxX &&
                point.getY() >= minY && point.getY() <= maxY;
    }
}
