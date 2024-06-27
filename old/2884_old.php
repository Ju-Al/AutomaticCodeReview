<?php

declare(strict_types=1);

/*
 * This file is part of Contao.
 *
 * (c) Leo Feyer
 *
 * @license LGPL-3.0-or-later
 */

namespace Contao\CoreBundle\Migration\Version408;

use Contao\CoreBundle\Framework\ContaoFramework;
use Contao\CoreBundle\Migration\AbstractMigration;
use Contao\CoreBundle\Migration\MigrationResult;
use Contao\File;
use Contao\StringUtil;
use Doctrine\DBAL\Connection;
use Symfony\Component\Filesystem\Filesystem;

/**
 * @internal
 */
class Version480Update extends AbstractMigration
{
    /**
     * @var Connection
     */
    private $connection;

    /**
     * @var Filesystem
     */
    private $filesystem;

    /**
     * @var ContaoFramework
     */
    private $framework;

    /**
     * @var string
     */
    private $projectDir;

    public function __construct(Connection $connection, Filesystem $filesystem, ContaoFramework $framework, string $projectDir)
    {
        $this->connection = $connection;
        $this->filesystem = $filesystem;
        $this->framework = $framework;
        $this->projectDir = $projectDir;
    }

    public function getName(): string
    {
        return 'Contao 4.8.0 Update';
    }

    public function shouldRun(): bool
    {
        return $this->shouldRunMediaelement()
            || $this->shouldRunSkipIfDimensionsMatch()
            || $this->shouldRunImportantPart()
            || $this->shouldRunMinKeywordLength()
            || $this->shouldRunContextLength()
            || $this->shouldRunDefaultImageDensities()
            || $this->shouldRunRememberMe();
    }

    public function run(): MigrationResult
    {
        $this->framework->initialize();

        if ($this->shouldRunMediaelement()) {
            $this->runMediaelement();
        }

        if ($this->shouldRunSkipIfDimensionsMatch()) {
            $this->runSkipIfDimensionsMatch();
        }

        if ($this->shouldRunImportantPart()) {
            $this->runImportantPart();
        }

        if ($this->shouldRunMinKeywordLength()) {
            $this->runMinKeywordLength();
        }

        if ($this->shouldRunContextLength()) {
            $this->runContextLength();
        }

        if ($this->shouldRunDefaultImageDensities()) {
            $this->runDefaultImageDensities();
        }

        if ($this->shouldRunRememberMe()) {
            $this->runRememberMe();
        }

        return $this->createResult(true);
    }

    public function shouldRunMediaelement(): bool
    {
        $schemaManager = $this->connection->getSchemaManager();

        if (!$schemaManager->tablesExist(['tl_layout'])) {
            return false;
        }

        $columns = $schemaManager->listTableColumns('tl_layout');

        if (!isset($columns['jquery'], $columns['scripts'])) {
            return false;
        }

        return (bool) $this->connection->query("
            SELECT EXISTS(
                SELECT id
                FROM tl_layout
                WHERE
                    jquery LIKE '%j_mediaelement%'
                    OR scripts LIKE '%js_mediaelement%'
            )
        ")->fetchColumn();
    }

    public function runMediaelement(): void
    {
        $statement = $this->connection->query('
            SELECT
                id, jquery, scripts
            FROM
                tl_layout
        ');

        // Remove the "j_mediaelement" and "js_mediaelement" templates
        while (false !== ($row = $statement->fetch(\PDO::FETCH_OBJ))) {
            if ($row->jquery) {
                $jquery = StringUtil::deserialize($row->jquery);

                if (\is_array($jquery) && false !== ($i = array_search('j_mediaelement', $jquery, true))) {
                    unset($jquery[$i]);

                    $stmt = $this->connection->prepare('
                        UPDATE
                            tl_layout
                        SET
                            jquery = :jquery
                        WHERE
                            id = :id
                    ');

                    $stmt->execute([':jquery' => serialize(array_values($jquery)), ':id' => $row->id]);
                }
            }

            if ($row->scripts) {
                $scripts = StringUtil::deserialize($row->scripts);

                if (\is_array($scripts) && false !== ($i = array_search('js_mediaelement', $scripts, true))) {
                    unset($scripts[$i]);

                    $stmt = $this->connection->prepare('
                        UPDATE
                            tl_layout
                        SET
                            scripts = :scripts
                        WHERE
                            id = :id
                    ');

                    $stmt->execute([':scripts' => serialize(array_values($scripts)), ':id' => $row->id]);
                }
            }
        }
    }

    public function shouldRunSkipIfDimensionsMatch(): bool
    {
        $schemaManager = $this->connection->getSchemaManager();

        if (!$schemaManager->tablesExist(['tl_image_size'])) {
            return false;
        }

        $columns = $schemaManager->listTableColumns('tl_image_size');

        return !isset($columns['skipifdimensionsmatch']);
    }

    public function runSkipIfDimensionsMatch(): void
    {
        $this->connection->query("
            ALTER TABLE
                tl_image_size
            ADD
                skipIfDimensionsMatch char(1) NOT NULL default ''
        ");

        // Enable the "skipIfDimensionsMatch" option for existing image sizes (backwards compatibility)
        $this->connection->query("
            UPDATE
                tl_image_size
            SET
                skipIfDimensionsMatch = '1'
        ");
    }

    public function shouldRunImportantPart(): bool
    {
        $schemaManager = $this->connection->getSchemaManager();

        if (!$schemaManager->tablesExist(['tl_files'])) {
            return false;
        }

        $columns = $schemaManager->listTableColumns('tl_files');

        if (
            !isset(
                $columns['path'],
                $columns['importantpartx'],
                $columns['importantparty'],
                $columns['importantpartwidth'],
                $columns['importantpartheight']
            )
        ) {
            return false;
        }

        return (bool) $this->connection->query('
            SELECT EXISTS(
                SELECT id
                FROM tl_files
                WHERE
                    importantPartX > 1
                    OR importantPartY > 1
                    OR importantPartWidth > 1
                    OR importantPartHeight > 1
            )
        ')->fetchColumn();
    }

    public function runImportantPart(): void
    {
        $this->connection->query('
            ALTER TABLE
                tl_files
            CHANGE
                importantPartX importantPartX DOUBLE PRECISION DEFAULT 0 NOT NULL,
            CHANGE
                importantPartY importantPartY DOUBLE PRECISION DEFAULT 0 NOT NULL,
            CHANGE
                importantPartWidth importantPartWidth DOUBLE PRECISION DEFAULT 0 NOT NULL,
            CHANGE
                importantPartHeight importantPartHeight DOUBLE PRECISION DEFAULT 0 NOT NULL
        ');

        $statement = $this->connection->query('
            SELECT
                id, path, importantPartX, importantPartY, importantPartWidth, importantPartHeight
            FROM
                tl_files
            WHERE
                importantPartWidth > 1 OR importantPartHeight > 1
        ');

        // Convert the important part to relative values as fractions
        while (false !== ($file = $statement->fetch(\PDO::FETCH_OBJ))) {
            if (!$this->filesystem->exists($this->projectDir.'/'.$file->path) || is_dir($this->projectDir.'/'.$file->path)) {
                continue;
            }

            $imageSize = (new File($file->path))->imageViewSize;

            if (empty($imageSize[0]) || empty($imageSize[1])) {
                continue;
            }

            $stmt = $this->connection->prepare('
                UPDATE
                    tl_files
                SET
                    importantPartX = :x,
                    importantPartY = :y,
                    importantPartWidth = :width,
                    importantPartHeight = :height
                WHERE
                    id = :id
            ');

            $stmt->execute([
                ':id' => $file->id,
                ':x' => $file->importantPartX / $imageSize[0],
                ':y' => $file->importantPartY / $imageSize[1],
                ':width' => $file->importantPartWidth / $imageSize[0],
                ':height' => $file->importantPartHeight / $imageSize[1],
            ]);
        }
    }

    public function shouldRunMinKeywordLength(): bool
    {
        $schemaManager = $this->connection->getSchemaManager();

        if (!$schemaManager->tablesExist(['tl_module'])) {
            return false;
        }

        $columns = $schemaManager->listTableColumns('tl_module');

        return !isset($columns['minkeywordlength']);
    }

    public function runMinKeywordLength(): void
    {
        $this->connection->query('
            ALTER TABLE
                tl_module
            ADD
                minKeywordLength smallint(5) unsigned NOT NULL default 4 AFTER contextLength
        ');

        // Disable the minimum keyword length for existing modules (backwards compatibility)
        $this->connection->query("
            UPDATE
                tl_module
            SET
                minKeywordLength = 0
            WHERE
                type = 'search'
        ");
    }

    public function shouldRunContextLength(): bool
    {
        $schemaManager = $this->connection->getSchemaManager();

        if (!$schemaManager->tablesExist(['tl_module'])) {
            return false;
        }

        $columns = $schemaManager->listTableColumns('tl_module');

        return isset($columns['contextlength'], $columns['totallength']);
    }

    public function runContextLength(): void
    {
        $this->connection->query("
            ALTER TABLE
                tl_module
            CHANGE
                contextLength contextLength varchar(64) NOT NULL default ''
        ");

        $statement = $this->connection->query("
            SELECT
                id, contextLength, totalLength
            FROM
                tl_module
            WHERE
                type = 'search'
        ");

        // Consolidate the search context fields
        while (false !== ($row = $statement->fetch(\PDO::FETCH_OBJ))) {
            $stmt = $this->connection->prepare('
                UPDATE
                    tl_module
                SET
                    contextLength = :context
                WHERE
                    id = :id
            ');

            $stmt->execute([
                ':id' => $row->id,
                ':context' => serialize([$row->contextLength, $row->totalLength]),
            ]);
        }

        $this->connection->query('ALTER TABLE tl_module DROP COLUMN totalLength');
    }

    public function shouldRunDefaultImageDensities(): bool
    {
        $schemaManager = $this->connection->getSchemaManager();

        if (!$schemaManager->tablesExist(['tl_layout', 'tl_theme'])) {
            return false;
        }

        $columnsLayout = $schemaManager->listTableColumns('tl_layout');
        $columnsTheme = $schemaManager->listTableColumns('tl_theme');

        return
            !isset($columnsLayout['defaultimagedensities'])
            && isset($columnsTheme['defaultimagedensities']);
    }

    public function runDefaultImageDensities(): void
    {
        $this->connection->query("
            ALTER TABLE
                tl_layout
            ADD
                defaultImageDensities varchar(255) NOT NULL default ''
        ");

        // Move the default image densities to the page layout
        $this->connection->query('
            UPDATE
                tl_layout l
            SET
                defaultImageDensities = (SELECT defaultImageDensities FROM tl_theme t WHERE t.id = l.pid)
        ');
    }

    public function shouldRunRememberMe(): bool
    {
        $schemaManager = $this->connection->getSchemaManager();

        if (!$schemaManager->tablesExist(['tl_remember_me'])) {
            return false;
        }

        $columns = $schemaManager->listTableColumns('tl_remember_me');

        return !isset($columns['id']);
    }

    public function runRememberMe(): void
    {
        // Since rememberme is broken in Contao 4.7 and there are no valid
        // cookies out there, we can simply drop the old table here and let the
        // install tool create the new one
        if ($this->connection->getSchemaManager()->tablesExist(['tl_remember_me'])) {
            $this->connection->query('DROP TABLE tl_remember_me');
        }
    }
}