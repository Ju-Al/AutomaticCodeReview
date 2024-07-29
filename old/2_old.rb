      end
    end

    def self.answer_string(host, answers)
      answers[host.name].map { |k,v| "#{k}=#{v}" }.join("\n")
    end

  end
end
