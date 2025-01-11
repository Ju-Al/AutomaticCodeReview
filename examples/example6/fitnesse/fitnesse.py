class Fitnesse:
    @staticmethod
    def testable_html(page_data, include_suite_setup):
        wiki_page = page_data.get_wiki_page()
        buffer = []
        if page_data.has_attribute("Test"):
            if include_suite_setup:
                suite_setup = PageCrawlerImpl.get_inherited_page(SuiteResponder.SUITE_SETUP_NAME, wiki_page)
                if suite_setup is not None:
                    page_path = suite_setup.get_page_crawler().get_full_path(suite_setup)
                    page_path_name = PathParser.render(page_path)
                    buffer.append(f"!include -setup .{page_path_name}\n")
            setup = PageCrawlerImpl.get_inherited_page("SetUp", wiki_page)
            if setup is not None:
                setup_path = wiki_page.get_page_crawler().get_full_path(setup)
                setup_path_name = PathParser.render(setup_path)
                buffer.append(f"!include -setup .{setup_path_name}\n")
        buffer.append(page_data.get_content())
        if page_data.has_attribute("Test"):
            teardown = PageCrawlerImpl.get_inherited_page("TearDown", wiki_page)
            if teardown is not None:
                teardown_path = wiki_page.get_page_crawler().get_full_path(teardown)
                teardown_path_name = PathParser.render(teardown_path)
                buffer.append(f"\n!include -teardown .{teardown_path_name}\n")
            if include_suite_setup:
                suite_teardown = PageCrawlerImpl.get_inherited_page(SuiteResponder.SUITE_TEARDOWN_NAME, wiki_page)
                if suite_teardown is not None:
                    page_path = suite_teardown.get_page_crawler().get_full_path(suite_teardown)
                    page_path_name = PathParser.render(page_path)
                    buffer.append(f"!include -teardown .{page_path_name}\n")
        page_data.set_content(''.join(buffer))
        return page_data.get_html()