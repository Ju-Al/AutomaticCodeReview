class PathsController < ApplicationController
    @paths = Path.all
  def show
    @path = Path.friendly.find(params[:id])
    @courses = decorated_courses
    @user = current_user
  end

  def index
    # The filter on here is temporary until the path is removed and the students
    #  enrolled in it are migrated to another path.
    @paths = Path.all.filter { |path| path.title != 'Front End Only' }
  end

  private

  def decorated_courses
    @path.courses.map { |course| CourseDecorator.new(course) }
  end
end
