<?php

namespace Tests\Wallabag\CoreBundle\Helper;

use Graby\Graby;
use Monolog\Handler\TestHandler;
use Monolog\Logger;
use PHPUnit\Framework\TestCase;
use Psr\Log\NullLogger;
use Symfony\Component\Validator\ConstraintViolation;
use Symfony\Component\Validator\ConstraintViolationList;
use Symfony\Component\Validator\Validator\RecursiveValidator;
use Wallabag\CoreBundle\Entity\Entry;
use Wallabag\CoreBundle\Helper\ContentProxy;
use Wallabag\CoreBundle\Helper\RuleBasedTagger;
use Wallabag\UserBundle\Entity\User;

class ContentProxyTest extends TestCase
{
    private $fetchingErrorMessage = 'wallabag can\'t retrieve contents for this article. Please <a href="http://doc.wallabag.org/en/user/errors_during_fetching.html#how-can-i-help-to-fix-that">troubleshoot this issue</a>.';

    public function testWithBadUrl()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => false,
                'title' => '',
                'url' => '',
                'content_type' => '',
                'language' => '',
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://user@:80');

        $this->assertSame('http://user@:80', $entry->getUrl());
        $this->assertEmpty($entry->getTitle());
        $this->assertSame($this->fetchingErrorMessage, $entry->getContent());
        $this->assertEmpty($entry->getPreviewPicture());
        $this->assertEmpty($entry->getMimetype());
        $this->assertEmpty($entry->getLanguage());
        $this->assertSame(0.0, $entry->getReadingTime());
        $this->assertNull($entry->getDomainName());
    }

    public function testWithEmptyContent()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => false,
                'title' => '',
                'url' => '',
                'content_type' => '',
                'language' => '',
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        $this->assertSame('http://0.0.0.0', $entry->getUrl());
        $this->assertEmpty($entry->getTitle());
        $this->assertSame($this->fetchingErrorMessage, $entry->getContent());
        $this->assertEmpty($entry->getPreviewPicture());
        $this->assertEmpty($entry->getMimetype());
        $this->assertEmpty($entry->getLanguage());
        $this->assertSame(0.0, $entry->getReadingTime());
        $this->assertSame('0.0.0.0', $entry->getDomainName());
    }

    public function testWithEmptyContentButOG()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => false,
                'title' => '',
                'url' => '',
                'content_type' => '',
                'language' => '',
                'status' => '',
                'open_graph' => [
                    'og_title' => 'my title',
                    'og_description' => 'desc',
                ],
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://domain.io');

        $this->assertSame('http://domain.io', $entry->getUrl());
        $this->assertSame('my title', $entry->getTitle());
        $this->assertSame($this->fetchingErrorMessage . '<p><i>But we found a short description: </i></p>desc', $entry->getContent());
        $this->assertEmpty($entry->getPreviewPicture());
        $this->assertEmpty($entry->getLanguage());
        $this->assertEmpty($entry->getHttpStatus());
        $this->assertEmpty($entry->getMimetype());
        $this->assertSame(0.0, $entry->getReadingTime());
        $this->assertSame('domain.io', $entry->getDomainName());
    }

    public function testWithContent()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => str_repeat('this is my content', 325),
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1',
                'content_type' => 'text/html',
                'language' => 'fr',
                'status' => '200',
                'open_graph' => [
                    'og_title' => 'my OG title',
                    'og_description' => 'OG desc',
                    'og_image' => 'http://3.3.3.3/cover.jpg',
                ],
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        $this->assertSame('http://1.1.1.1', $entry->getUrl());
        $this->assertSame('this is my title', $entry->getTitle());
        $this->assertContains('this is my content', $entry->getContent());
        $this->assertSame('http://3.3.3.3/cover.jpg', $entry->getPreviewPicture());
        $this->assertSame('text/html', $entry->getMimetype());
        $this->assertSame('fr', $entry->getLanguage());
        $this->assertSame('200', $entry->getHttpStatus());
        $this->assertSame(4.0, $entry->getReadingTime());
        $this->assertSame('1.1.1.1', $entry->getDomainName());
    }

    public function testWithContentAndNoOgImage()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => str_repeat('this is my content', 325),
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1',
                'content_type' => 'text/html',
                'language' => 'fr',
                'status' => '200',
                'open_graph' => [
                    'og_title' => 'my OG title',
                    'og_description' => 'OG desc',
                    'og_image' => null,
                ],
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        $this->assertSame('http://1.1.1.1', $entry->getUrl());
        $this->assertSame('this is my title', $entry->getTitle());
        $this->assertContains('this is my content', $entry->getContent());
        $this->assertNull($entry->getPreviewPicture());
        $this->assertSame('text/html', $entry->getMimetype());
        $this->assertSame('fr', $entry->getLanguage());
        $this->assertSame('200', $entry->getHttpStatus());
        $this->assertSame(4.0, $entry->getReadingTime());
        $this->assertSame('1.1.1.1', $entry->getDomainName());
    }

    public function testWithContentAndBadLanguage()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $validator = $this->getValidator(false);
        $validator->expects($this->once())
            ->method('validate')
            ->willReturn(new ConstraintViolationList([new ConstraintViolation('oops', 'oops', [], 'oops', 'language', 'dontexist')]));

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => str_repeat('this is my content', 325),
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1',
                'content_type' => 'text/html',
                'language' => 'dontexist',
                'status' => '200',
            ]);

        $proxy = new ContentProxy($graby, $tagger, $validator, $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        $this->assertSame('http://1.1.1.1', $entry->getUrl());
        $this->assertSame('this is my title', $entry->getTitle());
        $this->assertContains('this is my content', $entry->getContent());
        $this->assertSame('text/html', $entry->getMimetype());
        $this->assertNull($entry->getLanguage());
        $this->assertSame('200', $entry->getHttpStatus());
        $this->assertSame(4.0, $entry->getReadingTime());
        $this->assertSame('1.1.1.1', $entry->getDomainName());
    }

    public function testWithContentAndBadOgImage()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $validator = $this->getValidator(false);
        $validator->expects($this->exactly(2))
            ->method('validate')
            ->will($this->onConsecutiveCalls(
                new ConstraintViolationList(),
                new ConstraintViolationList([new ConstraintViolation('oops', 'oops', [], 'oops', 'url', 'https://')])
            ));

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => str_repeat('this is my content', 325),
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1',
                'content_type' => 'text/html',
                'language' => 'fr',
                'status' => '200',
                'open_graph' => [
                    'og_title' => 'my OG title',
                    'og_description' => 'OG desc',
                    'og_image' => 'https://',
                ],
            ]);

        $proxy = new ContentProxy($graby, $tagger, $validator, $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        $this->assertSame('http://1.1.1.1', $entry->getUrl());
        $this->assertSame('this is my title', $entry->getTitle());
        $this->assertContains('this is my content', $entry->getContent());
        $this->assertNull($entry->getPreviewPicture());
        $this->assertSame('text/html', $entry->getMimetype());
        $this->assertSame('fr', $entry->getLanguage());
        $this->assertSame('200', $entry->getHttpStatus());
        $this->assertSame(4.0, $entry->getReadingTime());
        $this->assertSame('1.1.1.1', $entry->getDomainName());
    }

    public function testWithForcedContent()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $proxy = new ContentProxy((new Graby()), $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage, true);
        $entry = new Entry(new User());
        $proxy->updateEntry(
            $entry,
            'http://0.0.0.0',
            [
                'html' => str_repeat('this is my content', 325),
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1',
                'content_type' => 'text/html',
                'language' => 'fr',
                'date' => '1395635872',
                'authors' => ['Jeremy', 'Nico', 'Thomas'],
                'all_headers' => [
                    'Cache-Control' => 'no-cache',
                ],
            ]
        );

        $this->assertSame('http://1.1.1.1', $entry->getUrl());
        $this->assertSame('this is my title', $entry->getTitle());
        $this->assertContains('this is my content', $entry->getContent());
        $this->assertSame('text/html', $entry->getMimetype());
        $this->assertSame('fr', $entry->getLanguage());
        $this->assertSame(4.0, $entry->getReadingTime());
        $this->assertSame('1.1.1.1', $entry->getDomainName());
        $this->assertSame('24/03/2014', $entry->getPublishedAt()->format('d/m/Y'));
        $this->assertContains('Jeremy', $entry->getPublishedBy());
        $this->assertContains('Nico', $entry->getPublishedBy());
        $this->assertContains('Thomas', $entry->getPublishedBy());
        $this->assertNotNull($entry->getHeaders(), 'Headers are stored, so value is not null');
        $this->assertContains('no-cache', $entry->getHeaders());
    }

    public function testWithForcedContentAndDatetime()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $logHandler = new TestHandler();
        $logger = new Logger('test', [$logHandler]);

        $proxy = new ContentProxy((new Graby()), $tagger, $this->getValidator(), $logger, $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry(
            $entry,
            'http://1.1.1.1',
            [
                'html' => str_repeat('this is my content', 325),
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1',
                'content_type' => 'text/html',
                'language' => 'fr',
                'date' => '2016-09-08T11:55:58+0200',
            ]
        );

        $this->assertSame('http://1.1.1.1', $entry->getUrl());
        $this->assertSame('this is my title', $entry->getTitle());
        $this->assertContains('this is my content', $entry->getContent());
        $this->assertSame('text/html', $entry->getMimetype());
        $this->assertSame('fr', $entry->getLanguage());
        $this->assertSame(4.0, $entry->getReadingTime());
        $this->assertSame('1.1.1.1', $entry->getDomainName());
        $this->assertSame('08/09/2016', $entry->getPublishedAt()->format('d/m/Y'));
    }

    public function testWithForcedContentAndBadDate()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $logger = new Logger('foo');
        $handler = new TestHandler();
        $logger->pushHandler($handler);

        $proxy = new ContentProxy((new Graby()), $tagger, $this->getValidator(), $logger, $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry(
            $entry,
            'http://1.1.1.1',
            [
                'html' => str_repeat('this is my content', 325),
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1',
                'content_type' => 'text/html',
                'language' => 'fr',
                'date' => '01 02 2012',
            ]
        );

        $this->assertSame('http://1.1.1.1', $entry->getUrl());
        $this->assertSame('this is my title', $entry->getTitle());
        $this->assertContains('this is my content', $entry->getContent());
        $this->assertSame('text/html', $entry->getMimetype());
        $this->assertSame('fr', $entry->getLanguage());
        $this->assertSame(4.0, $entry->getReadingTime());
        $this->assertSame('1.1.1.1', $entry->getDomainName());
        $this->assertNull($entry->getPublishedAt());

        $records = $handler->getRecords();

        $this->assertCount(1, $records);
        $this->assertContains('Error while defining date', $records[0]['message']);
    }

    public function testTaggerThrowException()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag')
            ->will($this->throwException(new \Exception()));

        $proxy = new ContentProxy((new Graby()), $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry(
            $entry,
            'http://1.1.1.1',
            [
                'html' => str_repeat('this is my content', 325),
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1',
                'content_type' => 'text/html',
                'language' => 'fr',
            ]
        );

        $this->assertCount(0, $entry->getTags());
    }

    public function dataForCrazyHtml()
    {
        return [
            'script and comment' => [
                '<strong>Script inside:</strong> <!--[if gte IE 4]><script>alert(\'lol\');</script><![endif]--><br />',
                'lol',
            ],
            'script' => [
                '<strong>Script inside:</strong><script>alert(\'lol\');</script>',
                'script',
            ],
        ];
    }

    /**
     * @dataProvider dataForCrazyHtml
     */
    public function testWithCrazyHtmlContent($html, $escapedString)
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $proxy = new ContentProxy((new Graby()), $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry(
            $entry,
            'http://1.1.1.1',
            [
                'html' => $html,
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1',
                'content_type' => 'text/html',
                'language' => 'fr',
                'status' => '200',
                'open_graph' => [
                    'og_title' => 'my OG title',
                    'og_description' => 'OG desc',
                    'og_image' => 'http://3.3.3.3/cover.jpg',
                ],
            ]
        );

        $this->assertSame('http://1.1.1.1', $entry->getUrl());
        $this->assertSame('this is my title', $entry->getTitle());
        $this->assertNotContains($escapedString, $entry->getContent());
        $this->assertSame('http://3.3.3.3/cover.jpg', $entry->getPreviewPicture());
        $this->assertSame('text/html', $entry->getMimetype());
        $this->assertSame('fr', $entry->getLanguage());
        $this->assertSame('200', $entry->getHttpStatus());
        $this->assertSame('1.1.1.1', $entry->getDomainName());
    }

    public function testWithImageAsContent()
    {
        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => '<p><img src="http://1.1.1.1/image.jpg" /></p>',
                'title' => 'this is my title',
                'url' => 'http://1.1.1.1/image.jpg',
                'content_type' => 'image/jpeg',
                'status' => '200',
                'open_graph' => [],
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        $this->assertSame('http://1.1.1.1/image.jpg', $entry->getUrl());
        $this->assertSame('this is my title', $entry->getTitle());
        $this->assertContains('http://1.1.1.1/image.jpg', $entry->getContent());
        $this->assertSame('http://1.1.1.1/image.jpg', $entry->getPreviewPicture());
        $this->assertSame('image/jpeg', $entry->getMimetype());
        $this->assertSame('200', $entry->getHttpStatus());
        $this->assertSame('1.1.1.1', $entry->getDomainName());
    {
        // You can use https://www.online-toolz.com/tools/text-hex-convertor.php to convert UTF-8 text <=> hex
        // See http://graphemica.com for more info about the characters
        // '😻ℤz' (U+1F63B or F09F98BB; U+2124 or E284A4; U+007A or 7A) in hexadecimal and UTF-8
        $actualTitle = $this->hexToStr('F09F98BB' . 'E284A4' . '7A');

        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => false,
                'title' => $actualTitle,
                'url' => '',
                'content_type' => 'text/html',
                'language' => '',
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        // '😻ℤz' (U+1F63B or F09F98BB; U+2124 or E284A4; U+007A or 7A) in hexadecimal and UTF-8
        $expectedTitle = 'F09F98BB' . 'E284A4' . '7A';
        $this->assertSame($expectedTitle, $this->strToHex($entry->getTitle()));
    }

    public function testWebsiteWithInvalidUTF8Title_removeInvalidCharacter()
    {
        // See http://graphemica.com for more info about the characters
        // 'a€b' (61;80;62) in hexadecimal and WINDOWS-1252 - but 80 is a invalid UTF-8 character.
        // The correct UTF-8 € character (U+20AC) is E282AC
        $actualTitle = $this->hexToStr('61' . '80' . '62');

        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => false,
                'title' => $actualTitle,
                'url' => '',
                'content_type' => 'text/html',
                'language' => '',
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        // 'ab' (61;62) because all invalid UTF-8 character (like 80) are removed
        $expectedTitle = '61' . '62';
        $this->assertSame($expectedTitle, $this->strToHex($entry->getTitle()));
    }

    public function testPdfWithUTF16BETitle_convertToUTF8()
    {
        // See http://graphemica.com for more info about the characters
        // '😻' (U+1F63B;D83DDE3B) in hexadecimal and as UTF16BE
        $actualTitle = $this->hexToStr('D83DDE3B');

        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => false,
                'title' => $actualTitle,
                'url' => '',
                'content_type' => 'application/pdf',
                'language' => '',
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        // '😻' (U+1F63B or F09F98BB) in hexadecimal and UTF-8
        $expectedTitle = 'F09F98BB';
        $this->assertSame($expectedTitle, $this->strToHex($entry->getTitle()));
    }

    public function testPdfWithUTF8Title_doNothing()
    {
        // See http://graphemica.com for more info about the characters
        // '😻' (U+1F63B;D83DDE3B) in hexadecimal and as UTF8
        $actualTitle = $this->hexToStr('F09F98BB');

        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => false,
                'title' => $actualTitle,
                'url' => '',
                'content_type' => 'application/pdf',
                'language' => '',
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        // '😻' (U+1F63B or F09F98BB) in hexadecimal and UTF-8
        $expectedTitle = 'F09F98BB';
        $this->assertSame($expectedTitle, $this->strToHex($entry->getTitle()));
    }

    public function testPdfWithWINDOWS1252Title_convertToUTF8()
    {
        // See http://graphemica.com for more info about the characters
        // '€' (80) in hexadecimal and WINDOWS-1252
        $actualTitle = $this->hexToStr('80');

        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => false,
                'title' => $actualTitle,
                'url' => '',
                'content_type' => 'application/pdf',
                'language' => '',
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        // '€' (U+20AC or E282AC) in hexadecimal and UTF-8
        $expectedTitle = 'E282AC';
        $this->assertSame($expectedTitle, $this->strToHex($entry->getTitle()));
    }

    public function testPdfWithInvalidCharacterInTitle_removeInvalidCharacter()
    {
        // See http://graphemica.com for more info about the characters
        // '😻ℤ�z' (U+1F63B or F09F98BB; U+2124 or E284A4; invalid character 81; U+007A or 7A) in hexadecimal and UTF-8
        // 0x81 is not a valid character for UTF16, UTF8 and WINDOWS-1252
        $actualTitle = $this->hexToStr('F09F98BB' . 'E284A4' . '81' . '7A');

        $tagger = $this->getTaggerMock();
        $tagger->expects($this->once())
            ->method('tag');

        $graby = $this->getMockBuilder('Graby\Graby')
            ->setMethods(['fetchContent'])
            ->disableOriginalConstructor()
            ->getMock();

        $graby->expects($this->any())
            ->method('fetchContent')
            ->willReturn([
                'html' => false,
                'title' => $actualTitle,
                'url' => '',
                'content_type' => 'application/pdf',
                'language' => '',
            ]);

        $proxy = new ContentProxy($graby, $tagger, $this->getValidator(), $this->getLogger(), $this->fetchingErrorMessage);
        $entry = new Entry(new User());
        $proxy->updateEntry($entry, 'http://0.0.0.0');

        // '😻ℤz' (U+1F63B or F09F98BB; U+2124 or E284A4; U+007A or 7A) in hexadecimal and UTF-8
        // the 0x81 (represented by �) is invalid for UTF16, UTF8 and WINDOWS-1252 and is removed
        $expectedTitle = 'F09F98BB' . 'E284A4' . '7A';
        $this->assertSame($expectedTitle, $this->strToHex($entry->getTitle()));
    }

    /**
     * https://stackoverflow.com/a/18506801
     * @param $string
     * @return string
     */
    function strToHex($string){
        $hex = '';
        for ($i=0; $i<strlen($string); $i++){
            $ord = ord($string[$i]);
            $hexCode = dechex($ord);
            $hex .= substr('0'.$hexCode, -2);
        }
        return strToUpper($hex);
    }

    /**
     * https://stackoverflow.com/a/18506801
     * @param $hex
     * @return string
     */
    function hexToStr($hex){
        $string='';
        for ($i=0; $i < strlen($hex)-1; $i+=2){
            $string .= chr(hexdec($hex[$i].$hex[$i+1]));
        }
        return $string;
    }

    private function getTaggerMock()
    {
        return $this->getMockBuilder(RuleBasedTagger::class)
            ->setMethods(['tag'])
            ->disableOriginalConstructor()
            ->getMock();
    }

    private function getLogger()
    {
        return new NullLogger();
    }

    private function getValidator($withDefaultMock = true)
    {
        $mock = $this->getMockBuilder(RecursiveValidator::class)
            ->setMethods(['validate'])
            ->disableOriginalConstructor()
            ->getMock();

        if ($withDefaultMock) {
            $mock->expects($this->any())
                ->method('validate')
                ->willReturn(new ConstraintViolationList());
        }

        return $mock;
    }
}
