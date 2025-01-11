from point import Point
from line import Line
from lineSegment import LineSegment

def main():
    p1 = Point(0, 0)
    p2 = Point(4, 4)
    test_point = Point(2, 2)

    line = Line(p1, p2)
    segment = LineSegment(p1, p2)

    print(f"Slope: {line.get_slope()}")
    print(f"Intercept: {line.get_intercept()}")
    print(f"Point on line: {line.is_on(test_point)}")
    print(f"Segment length: {segment.get_length()}")
    print(f"Point on segment: {segment.is_on(test_point)}")

if __name__ == "__main__":
    main()