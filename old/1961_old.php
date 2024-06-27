<?php
/**
 * Holds Controller
 *
 * PHP version 7
 *
 * Copyright (C) Villanova University 2010.
 * Copyright (C) The National Library of Finland 2021.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 2,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * @category VuFind
 * @package  Controller
 * @author   Demian Katz <demian.katz@villanova.edu>
 * @author   Ere Maijala <ere.maijala@helsinki.fi>
 * @license  http://opensource.org/licenses/gpl-2.0.php GNU General Public License
 * @link     https://vufind.org Main Site
 */
namespace VuFind\Controller;

use Laminas\ServiceManager\ServiceLocatorInterface;
use VuFind\Exception\ILS as ILSException;

/**
 * Controller for the user holds area.
 *
 * @category VuFind
 * @package  Controller
 * @author   Demian Katz <demian.katz@villanova.edu>
 * @author   Ere Maijala <ere.maijala@helsinki.fi>
 * @license  http://opensource.org/licenses/gpl-2.0.php GNU General Public License
 * @link     https://vufind.org Main Site
 */
class HoldsController extends AbstractBase
{
    /**
     * Hold update results container
     *
     * @var \Laminas\Session\Container
     */
    protected $updateResultsContainer;

    /**
     * Constructor
     *
     * @param ServiceLocatorInterface    $sm  Service locator
     * @param \Laminas\Session\Container $urc A session container for hold update
     * results
     */
    public function __construct(ServiceLocatorInterface $sm,
        \Laminas\Session\Container $urc
    ) {
        $this->serviceLocator = $sm;
        $this->updateResultsContainer = $urc;
    }

    /**
     * Send list of holds to view
     *
     * @return mixed
     */
    public function listAction()
    {
        // Stop now if the user does not have valid catalog credentials available:
        if (!is_array($patron = $this->catalogLogin())) {
            return $patron;
        }

        // Connect to the ILS:
        $catalog = $this->getILS();

        // Process cancel requests if necessary:
        $cancelStatus = $catalog->checkFunction('cancelHolds', compact('patron'));
        $view = $this->createViewModel();
        $view->cancelResults = $cancelStatus
            ? $this->holds()->cancelHolds($catalog, $patron) : [];
        // If we need to confirm
        if (!is_array($view->cancelResults)) {
            return $view->cancelResults;
        }

        // Process any update request results stored in the session:
        $holdUpdateResults = $this->updateResultsContainer->results ?? null;
        if ($holdUpdateResults) {
            $view->updateResults = $holdUpdateResults;
            $this->updateResultsContainer->results = null;
        }
        // Process update requests if necessary:
        if ($this->params()->fromPost('updateSelected')) {
            $details = $this->params()->fromPost('selectedIDS');
            if (empty($details)) {
                $this->flashMessenger()->addErrorMessage('hold_empty_selection');
                if ($this->inLightbox()) {
                    return $this->getRefreshResponse();
                }
            } else {
                return $this->forwardTo('Holds', 'Edit');
            }
        }

        // By default, assume we will not need to display a cancel or update form:
        $view->cancelForm = false;
        $view->updateForm = false;

        // Get held item details:
        $result = $catalog->getMyHolds($patron);
        $driversNeeded = [];
        $this->holds()->resetValidation();
        $holdConfig = $catalog->checkFunction('Holds', compact('patron'));
        foreach ($result as $current) {
            // Add cancel details if appropriate:
            $current = $this->holds()->addCancelDetails(
                $catalog, $current, $cancelStatus
            );
            if ($cancelStatus && $cancelStatus['function'] != "getCancelHoldLink"
                && isset($current['cancel_details'])
            ) {
                // Enable cancel form if necessary:
                $view->cancelForm = true;
            }

            // Add update details if appropriate
            if (!empty($holdConfig['updateFields'])) {
                $current = $this->holds()->addUpdateDetails(
                    $catalog,
                    $current,
                    $holdConfig['updateFields']
                );
                if (isset($current['updateDetails'])) {
                    $view->updateForm = true;
                }
            }

            $driversNeeded[] = $current;
        }

        // Get List of PickUp Libraries based on patron's home library
        try {
            $view->pickup = $catalog->getPickUpLocations($patron);
        } catch (\Exception $e) {
            // Do nothing; if we're unable to load information about pickup
            // locations, they are not supported and we should ignore them.
        }

        $view->recordList = $this->ilsRecords()->getDrivers($driversNeeded);
        $view->accountStatus = $this->ilsRecords()
            ->collectRequestStats($view->recordList);
        return $view;
    }

    /**
     * Edit holds
     *
     * @return mixed
     */
    public function editAction()
    {
        // Stop now if the user does not have valid catalog credentials available:
        if (!is_array($patron = $this->catalogLogin())) {
            return $patron;
        }

        // Connect to the ILS:
        $catalog = $this->getILS();

        $holdConfig = $catalog->checkFunction('Holds', compact('patron'));
        $details = $this->params()->fromPost('selectedIDS');
        if (empty($holdConfig['updateFields']) || empty($details)) {
            // Shouldn't be here. Redirect to holds
            return $this->redirect()->toRoute('holds-list');
        }
        // If the user input contains a value not found in the session
        // legal list, something has been tampered with -- abort the process.
        if (array_diff($details, $this->holds()->getValidIds())) {
            $this->flashMessenger()
                ->addErrorMessage('error_inconsistent_parameters');
            return $this->redirect()->toRoute('holds-list');
        }

        // Get list of pickup locations based on the first selected hold. This may
        // not be perfect as pickup locations may differ per request, but it's the
        // best we can do.
        $holds = $catalog->getMyHolds($patron);
        $firstDetails = reset($details);
        $pickupLocations = [];
        foreach ($holds as $hold) {
            if ((string)$catalog->getUpdateHoldDetails($hold) === $firstDetails) {
                try {
                    $pickupLocations = $catalog->getPickUpLocations($patron, $hold);
                } catch (ILSException $e) {
                    $this->flashMessenger()
                        ->addErrorMessage('ils_connection_failed');
                }
                break;
            }
        }

        $gatheredDetails = $this->params()->fromPost('gatheredDetails', []);
        if ($this->params()->fromPost('updateHolds')) {
            $validPickup = true;
            $selectedPickupLocation = $gatheredDetails['pickUpLocation'] ?? '';
            if ('' !== $selectedPickupLocation) {
                $validPickup = $this->holds()->validatePickUpInput(
                    $selectedPickupLocation,
                    $holdConfig['updateFields'],
                    $pickupLocations
                );
            }
            $dateValidationResults = $this->holds()->validateDates(
                $gatheredDetails['startDate'] ?? null,
                $gatheredDetails['requiredBy'] ?? null,
                $holdConfig['updateFields']
            );
            if (in_array('frozenUntil', $holdConfig['updateFields'])) {
                $frozenUntilValidationResults = $this->holds()->validateFrozenUntil(
                    $gatheredDetails['frozenUntil'] ?? null,
                    $holdConfig['updateFields']
                );
                $dateValidationResults['errors'] = array_unique(
                    array_merge(
                        $dateValidationResults['errors'],
                        $frozenUntilValidationResults['errors']
                    )
                );
            }
            if (!$validPickup) {
                $this->flashMessenger()->addErrorMessage('hold_invalid_pickup');
            }
            foreach ($dateValidationResults['errors'] as $msg) {
                $this->flashMessenger()->addErrorMessage($msg);
            }
            if ($validPickup && !$dateValidationResults['errors']) {
                $updateFields = [];
                if ($selectedPickupLocation !== '') {
                    $updateFields['pickUpLocation'] = $selectedPickupLocation;
                }
                if ($gatheredDetails['startDate'] ?? '' !== '') {
                    $updateFields['startDate'] = $gatheredDetails['startDate'];
                    $updateFields['startDateTS']
                        = $dateValidationResults['startDateTS'];
                }
                if (($gatheredDetails['requiredBy'] ?? '') !== '') {
                    $updateFields['requiredBy'] = $gatheredDetails['requiredBy'];
                    $updateFields['requiredByTS']
                        = $dateValidationResults['requiredByTS'];
                }
                if (($gatheredDetails['frozen'] ?? '') !== '') {
                    $updateFields['frozen'] = $gatheredDetails['frozen'] === '1';
                    if (($gatheredDetails['frozenUntil']) ?? '' !== '') {
                        $updateFields['frozenUntil']
                            = $gatheredDetails['frozenUntil'];
                        $updateFields['frozenUntilTS']
                            = $frozenUntilValidationResults['frozenUntilTS'];
                    }
                }

                $results = $catalog->updateHolds($patron, $details, $updateFields);
                $successful = 0;
                $failed = 0;
                foreach ($results as $result) {
                    if ($result['success']) {
                        ++$successful;
                    } else {
                        ++$failed;
                    }
                }
                // Store results in the session so that they can be displayed when
                // the user is redirected back to the holds list:
                $this->updateResultsContainer->results = $results;
                if ($successful) {
                    $msg = $this->translate(
                        'hold_edit_success_items',
                        ['%%count%%' => $successful]
                    );
                    $this->flashMessenger()->addSuccessMessage($msg);
                }
                if ($failed) {
                    $msg = $this->translate(
                        'hold_edit_failed_items',
                        ['%%count%%' => $failed]
                    );
                    $this->flashMessenger()->addErrorMessage($msg);
                }
                return $this->inLightbox()
                    ? $this->getRefreshResponse()
                    : $this->redirect()->toRoute('holds-list');
            }
        }

        $view = $this->createViewModel(
            [
                'selectedIDS' => $details,
                'fields' => $holdConfig['updateFields'],
                'gatheredDetails' => $gatheredDetails,
                'pickupLocations' => $pickupLocations,
            ]
        );

        return $view;
    }
}
