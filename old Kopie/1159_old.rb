# frozen_string_literal: true

class ReportCallbacks
  def after_create(report)
    if report.user.reports.count == 1
      send_first_report_notification(report)
      if report.user.reports.count == 1
        send_first_report_notification(report)
      end
    end

    if report.user.trainee?
      report.user.company.advisers.each do |adviser|
        send_trainee_report_notification(report, adviser)
        create_advisers_watch(report, adviser)
      end
    end
  end

  def after_update(report)
    unless report.wip
      if report.user.reports.count == 1
        send_first_report_notification(report)
      end
    end
  end

  private
    def send_first_report_notification(report)
      receiver_list = User.where(retired_on: nil)
      receiver_list.each do |receiver|
        if report.sender != receiver
          NotificationFacade.first_report(report, receiver)
        end
      end
    end

    def send_trainee_report_notification(report, receiver)
      NotificationFacade.trainee_report(report, receiver)
    end

    def create_advisers_watch(report, adviser)
      Watch.create!(user: adviser, watchable: report)
    end
end
