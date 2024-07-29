    pictures = fetch_pictures(crop.name)
    pictures.each do |picture|
      data = picture.fetch('attributes')
      Rails.logger.debug(data)
      next unless data.fetch('image_url').start_with? 'http'
      next if Photo.find_by(source_id: picture.fetch('id'), source: 'openfarm')

      photo = Photo.new(source_id: picture.fetch('id'), source: 'openfarm')
      photo.owner = @cropbot
      photo.thumbnail_url = data.fetch('thumbnail_url')
      photo.fullsize_url = data.fetch('image_url')
      photo.title = 'Open Farm photo'
      photo.license_name = 'No rights reserved'
      photo.link_url = "https://openfarm.cc/en/crops/#{name_to_slug(crop.name)}"
      if photo.valid?
        photo.save

        PhotoAssociation.find_or_create_by! photo: photo, photographable: crop
        Rails.logger.debug "\t saved photo #{photo.id} #{photo.source_id}"
      else
        Photo.where(thumbnail_url: photo.thumbnail_url).each do |p|
          Rails.logger.warn p
        end
        Photo.where(fullsize_url: photo.fullsize_url).each do |p|
          Rails.logger.warn p
        end
        Rails.logger.warn "Photo not valid"
      end
    end
  end

  def fetch(name)
    conn.get("crops/#{name_to_slug(name)}.json").body
  rescue NoMethodError
    Rails.logger.debug "error fetching crop"
    Rails.logger.debug "BODY: "
    Rails.logger.debug body
    Rails.logger.debug " =================== "
  end

  def name_to_slug(name)
    CGI.escape(name.gsub(' ', '-').downcase)
  end

  def fetch_all(page)
    conn.get("crops.json?page=#{page}").body.fetch('data', {})
  end

  def fetch_pictures(name)
    body = conn.get("crops/#{name_to_slug(name)}/pictures.json").body
    body.fetch('data', false)
  rescue StandardError
    Rails.logger.debug "Error fetching photos"
    Rails.logger.debug []
  end

  private

  def conn
    Faraday.new BASE do |conn|
      conn.response :json, content_type: /\bjson$/
      conn.adapter Faraday.default_adapter
    end
  end
end
