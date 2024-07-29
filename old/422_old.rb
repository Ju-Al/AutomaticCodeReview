          comments_count: 2, # TODO
          lines_count: 9, # TODO
          snippet: snippet, # TODO
          highlightjs_language: "csharp" # TODO
        },
        layout: false
    end

    private
    attr_reader :solution

    def snippet
      '
public class Year
{
  public static bool IsLeap(int year)
  {
      if (year % 4 != 0) return false
      if (year % 100 == 0 && year % 400) return false
      return true;
  }
}
      '.strip
    end
  end
end
