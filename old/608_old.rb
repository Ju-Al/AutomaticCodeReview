
  def validate(record)
    
    if record.image_url.present?

      validation_regex = %r{\.(png|jpg|jpeg)$}i
      if validation_regex.match(record.image_url) == nil
        record.errors[:image_url] = 'Invalid format. Image must be png, jpg, or jpeg.'
      end

      unless is_whitelisted?(record.image_url) 
        record.errors[:image_url] = 'Invalid image url. Image provider not found in provider whitelist.'
      end
    end
  end

  private


  def is_whitelisted?(url)
    IMAGE_HOST_WHITELIST.any? {|whitelist_item| url.match /#{Regexp.escape(whitelist_item)}/ }
  end
end