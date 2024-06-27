<?php

namespace Wallabag\CoreBundle\Repository;

use Doctrine\ORM\EntityRepository;
use Pagerfanta\Adapter\DoctrineORMAdapter;
use Pagerfanta\Pagerfanta;
use Wallabag\CoreBundle\Entity\Tag;

class EntryRepository extends EntityRepository
{
    /**
     * Return a query builder to used by other getBuilderFor* method.
     *
     * @param int $userId
     *
     * @return QueryBuilder
     */
    private function getBuilderByUser($userId)
    {
        return $this->createQueryBuilder('e')
            ->leftJoin('e.user', 'u')
            ->andWhere('u.id = :userId')->setParameter('userId', $userId)
            ->orderBy('e.id', 'desc')
        ;
    }

    /**
     * Retrieves all entries for a user.
     *
     * @param int $userId
     *
     * @return QueryBuilder
     */
    public function getBuilderForAllByUser($userId)
    {
        return $this
            ->getBuilderByUser($userId)
        ;
    }

    /**
     * Retrieves unread entries for a user.
     *
     * @param int $userId
     *
     * @return QueryBuilder
     */
    public function getBuilderForUnreadByUser($userId)
    {
        return $this
            ->getBuilderByUser($userId)
            ->andWhere('e.isArchived = false')
        ;
    }

    /**
     * Retrieves read entries for a user.
     *
     * @param int $userId
     *
     * @return QueryBuilder
     */
    public function getBuilderForArchiveByUser($userId)
    {
        return $this
            ->getBuilderByUser($userId)
            ->andWhere('e.isArchived = true')
        ;
    }

    /**
     * Retrieves starred entries for a user.
     *
     * @param int $userId
     *
     * @return QueryBuilder
     */
    public function getBuilderForStarredByUser($userId)
    {
        return $this
            ->getBuilderByUser($userId)
            ->andWhere('e.isStarred = true')
        ;
    }

    /**
     * Find Entries.
     *
     * @param int    $userId
     * @param bool   $isArchived
     * @param bool   $isStarred
     * @param string $sort
     * @param string $order
     *
     * @return array
     */
    public function findEntries($userId, $isArchived = null, $isStarred = null, $sort = 'created', $order = 'ASC')
    {
        $qb = $this->createQueryBuilder('e')
            ->where('e.user =:userId')->setParameter('userId', $userId);

        if (null !== $isArchived) {
            $qb->andWhere('e.isArchived =:isArchived')->setParameter('isArchived', (bool) $isArchived);
        }

        if (null !== $isStarred) {
            $qb->andWhere('e.isStarred =:isStarred')->setParameter('isStarred', (bool) $isStarred);
        }

        if ('created' === $sort) {
            $qb->orderBy('e.id', $order);
        } elseif ('updated' === $sort) {
            $qb->orderBy('e.updatedAt', $order);
        }

        $pagerAdapter = new DoctrineORMAdapter($qb);

        return new Pagerfanta($pagerAdapter);
    }

    /**
     * Fetch an entry with a tag. Only used for tests.
     *
     * @return Entry
     */
    public function findOneWithTags($userId)
    {
        $qb = $this->createQueryBuilder('e')
            ->innerJoin('e.tags', 't')
            ->innerJoin('e.user', 'u')
            ->addSelect('t', 'u')
            ->where('e.user=:userId')->setParameter('userId', $userId)
        ;

        return $qb->getQuery()->getResult();
    }

    /**
     * Find distinct language for a given user.
     * Used to build the filter language list.
     *
     * @param int $userId User id
     *
     * @return array
     */
    public function findDistinctLanguageByUser($userId)
    {
        $results = $this->createQueryBuilder('e')
            ->select('e.language')
            ->where('e.user = :userId')->setParameter('userId', $userId)
            ->andWhere('e.language IS NOT NULL')
            ->groupBy('e.language')
            ->orderBy('e.language', ' ASC')
            ->getQuery()
            ->getResult();

        $languages = array();
        foreach ($results as $result) {
            $languages[$result['language']] = $result['language'];
        }

        return $languages;
    }

    /**
     * Used only in test case to get the right entry associated to the right user.
     *
     * @param string $username
     *
     * @return Entry
     */
    public function findOneByUsernameAndNotArchived($username)
    {
        return $this->createQueryBuilder('e')
            ->leftJoin('e.user', 'u')
            ->where('u.username = :username')->setParameter('username', $username)
            ->andWhere('e.isArchived = false')
            ->setMaxResults(1)
            ->getQuery()
            ->getSingleResult();
    }

    /**
     * Remove a tag from all user entries.
     *
     * We need to loop on each entry attached to the given tag to remove it, since Doctrine doesn't know EntryTag entity because it's a ManyToMany relation.
     * It could be faster with one query but I don't know how to retrieve the table name `entry_tag` which can have a prefix.
     *
     * @param int $userId
     * @param Tag $tag
     */
    public function removeTag($userId, Tag $tag)
    {
        $entries = $this->getBuilderByUser($userId)
            ->innerJoin('e.tags', 't')
            ->andWhere('t.id = :tagId')->setParameter('tagId', $tag->getId())
            ->getQuery()
            ->getResult();

        foreach ($entries as $entry) {
            $entry->removeTag($tag);
        }

        $this->getEntityManager()->flush();

        // An other solution can be to use raw query but I can't find a way to retrieve the `entry_tag` table name since it can be prefixed....
        // $sql = 'DELETE et FROM entry_tag et WHERE et.entry_id IN ( SELECT e.id FROM entry e WHERE e.user_id = :userId ) AND et.tag_id = :tagId';
        // $stmt = $this->getEntityManager()->getConnection()->prepare($sql);
        // $stmt->execute([
        //     'userId' => $userId,
        //     'tagId' => $tag->getId(),
        // ]);
    }

    /**
     * Find all entries that are attached to a give tag id.
     *
     * @param int $userId
     * @param int $tagId
     *
     * @return array
     */
    public function findAllByTagId($userId, $tagId)
    {
        return $this->getBuilderByUser($userId)
            ->innerJoin('e.tags', 't')
            ->andWhere('t.id = :tagId')->setParameter('tagId', $tagId)
            ->getQuery()
            ->getResult();
    }
}
