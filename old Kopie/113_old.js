/**
 * @license
 * Visual Blocks Editor
 *
 * Copyright 2014 Google Inc.
 * https://developers.google.com/blockly/
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * @fileoverview Object representing a workspace rendered as SVG.
 * @author fraser@google.com (Neil Fraser)
 */
'use strict';

goog.provide('Blockly.WorkspaceSvg');

// TODO(scr): Fix circular dependencies
// goog.require('Blockly.Block');
goog.require('Blockly.ScrollbarPair');
goog.require('Blockly.Trashcan');
goog.require('Blockly.ZoomControls');
goog.require('Blockly.Workspace');
goog.require('Blockly.Xml');

goog.require('goog.dom');
goog.require('goog.math.Coordinate');
goog.require('goog.userAgent');


/**
 * Class for a workspace.  This is an onscreen area with optional trashcan,
 * scrollbars, bubbles, and dragging.
 * @param {!Object} options Dictionary of options.
 * @extends {Blockly.Workspace}
 * @constructor
 */
Blockly.WorkspaceSvg = function(options) {
  Blockly.WorkspaceSvg.superClass_.constructor.call(this, options);
  this.getMetrics = options.getMetrics;
  this.setMetrics = options.setMetrics;

  /** @type {!Function} */
  this.audioCallback_ = null;

  Blockly.ConnectionDB.init(this);
};
goog.inherits(Blockly.WorkspaceSvg, Blockly.Workspace);

/**
 * Svg workspaces are user-visible (as opposed to a headless workspace).
 * @type {boolean} True if visible.  False if headless.
 */
Blockly.WorkspaceSvg.prototype.rendered = true;

/**
 * Is this workspace the surface for a flyout?
 * @type {boolean}
 */
Blockly.WorkspaceSvg.prototype.isFlyout = false;

/**
 * Is this workspace currently being dragged around?
 * @type {boolean}
 */
Blockly.WorkspaceSvg.prototype.isScrolling = false;

/**
 * Current horizontal scrolling offset.
 * @type {number}
 */
Blockly.WorkspaceSvg.prototype.scrollX = 0;

/**
 * Current vertical scrolling offset.
 * @type {number}
 */
Blockly.WorkspaceSvg.prototype.scrollY = 0;

/**
 * Horizontal scroll value when scrolling started.
 * @type {number}
 */
Blockly.WorkspaceSvg.prototype.startScrollX = 0;

/**
 * Vertical scroll value when scrolling started.
 * @type {number}
 */
Blockly.WorkspaceSvg.prototype.startScrollY = 0;

/**
 * Horizontal distance from mouse to object being dragged.
 * @type {number}
 * @private
 */
Blockly.WorkspaceSvg.prototype.dragDeltaX_ = 0;

/**
 * Vertical distance from mouse to object being dragged.
 * @type {number}
 * @private
 */
Blockly.WorkspaceSvg.prototype.dragDeltaY_ = 0;

/**
 * Current scale.
 * @type {number}
 */
Blockly.WorkspaceSvg.prototype.scale = 1;

/**
 * The workspace's trashcan (if any).
 * @type {Blockly.Trashcan}
 */
Blockly.WorkspaceSvg.prototype.trashcan = null;

/**
 * This workspace's scrollbars, if they exist.
 * @type {Blockly.ScrollbarPair}
 */
Blockly.WorkspaceSvg.prototype.scrollbar = null;

/**
 * Create the workspace DOM elements.
 * @param {string=} opt_backgroundClass Either 'blocklyMainBackground' or
 *     'blocklyMutatorBackground'.
 * @return {!Element} The workspace's SVG group.
 */
Blockly.WorkspaceSvg.prototype.createDom = function(opt_backgroundClass) {
  /**
   * <g class="blocklyWorkspace">
   *   <rect class="blocklyMainBackground" height="100%" width="100%"></rect>
   *   [Trashcan and/or flyout may go here]
   *   <g class="blocklyBlockCanvas"></g>
   *   <g class="blocklyBubbleCanvas"></g>
   *   [Scrollbars may go here]
   * </g>
   * @type {SVGElement}
   */
  this.svgGroup_ = Blockly.createSvgElement('g',
      {'class': 'blocklyWorkspace'}, null);
  if (opt_backgroundClass) {
    /** @type {SVGElement} */
    this.svgBackground_ = Blockly.createSvgElement('rect',
        {'height': '100%', 'width': '100%', 'class': opt_backgroundClass},
        this.svgGroup_);
    if (opt_backgroundClass == 'blocklyMainBackground') {
      this.svgBackground_.style.fill =
          'url(#' + this.options.gridPattern.id + ')';
    }
  }
  /** @type {SVGElement} */
  this.svgBlockCanvas_ = Blockly.createSvgElement('g',
      {'class': 'blocklyBlockCanvas'}, this.svgGroup_, this);
  /** @type {SVGElement} */
  this.svgBubbleCanvas_ = Blockly.createSvgElement('g',
      {'class': 'blocklyBubbleCanvas'}, this.svgGroup_, this);
  var bottom = Blockly.Scrollbar.scrollbarThickness;
  if (this.options.hasTrashcan) {
    bottom = this.addTrashcan_(bottom);
  }
  if (this.options.zoomOptions && this.options.zoomOptions.controls) {
    bottom = this.addZoomControls_(bottom);
  }
  Blockly.bindEvent_(this.svgGroup_, 'mousedown', this, this.onMouseDown_);
  var thisWorkspace = this;
  Blockly.bindEvent_(this.svgGroup_, 'touchstart', null,
                     function(e) {Blockly.longStart_(e, thisWorkspace);});
  if (this.options.zoomOptions && this.options.zoomOptions.wheel) {
    // Mouse-wheel.
    Blockly.bindEvent_(this.svgGroup_, 'wheel', this, this.onMouseWheel_);
  }

  // Determine if there needs to be a category tree, or a simple list of
  // blocks.  This cannot be changed later, since the UI is very different.
  if (this.options.hasCategories) {
    this.toolbox_ = new Blockly.Toolbox(this);
  } else if (this.options.languageTree) {
    this.addFlyout_();
  }
  this.updateGridPattern_();
  return this.svgGroup_;
};

/**
 * Dispose of this workspace.
 * Unlink from all DOM elements to prevent memory leaks.
 */
Blockly.WorkspaceSvg.prototype.dispose = function() {
  // Stop rerendering.
  this.rendered = false;
  Blockly.WorkspaceSvg.superClass_.dispose.call(this);
  if (this.svgGroup_) {
    goog.dom.removeNode(this.svgGroup_);
    this.svgGroup_ = null;
  }
  this.svgBlockCanvas_ = null;
  this.svgBubbleCanvas_ = null;
  if (this.toolbox_) {
    this.toolbox_.dispose();
    this.toolbox_ = null;
  }
  if (this.flyout_) {
    this.flyout_.dispose();
    this.flyout_ = null;
  }
  if (this.trashcan) {
    this.trashcan.dispose();
    this.trashcan = null;
  }
  if (this.scrollbar) {
    this.scrollbar.dispose();
    this.scrollbar = null;
  }
  if (this.zoomControls_) {
    this.zoomControls_.dispose();
    this.zoomControls_ = null;
  }
  if (!this.options.parentWorkspace) {
    // Top-most workspace.  Dispose of the SVG too.
    goog.dom.removeNode(this.getParentSvg());
  }
};

/**
 * Obtain a newly created block.
 * @param {?string} prototypeName Name of the language object containing
 *     type-specific functions for this block.
 * @param {=string} opt_id Optional ID.  Use this ID if provided, otherwise
 *     create a new id.
 * @return {!Blockly.BlockSvg} The created block.
 */
Blockly.WorkspaceSvg.prototype.newBlock = function(prototypeName, opt_id) {
  return new Blockly.BlockSvg(this, prototypeName, opt_id);
};

/**
 * Add a trashcan.
 * @param {number} bottom Distance from workspace bottom to bottom of trashcan.
 * @return {number} Distance from workspace bottom to the top of trashcan.
 * @private
 */
Blockly.WorkspaceSvg.prototype.addTrashcan_ = function(bottom) {
  /** @type {Blockly.Trashcan} */
  this.trashcan = new Blockly.Trashcan(this);
  var svgTrashcan = this.trashcan.createDom();
  this.svgGroup_.insertBefore(svgTrashcan, this.svgBlockCanvas_);
  return this.trashcan.init(bottom);
};

/**
 * Add zoom controls.
 * @param {number} bottom Distance from workspace bottom to bottom of controls.
 * @return {number} Distance from workspace bottom to the top of controls.
 * @private
 */
Blockly.WorkspaceSvg.prototype.addZoomControls_ = function(bottom) {
  /** @type {Blockly.ZoomControls} */
  this.zoomControls_ = new Blockly.ZoomControls(this);
  var svgZoomControls = this.zoomControls_.createDom();
  this.svgGroup_.appendChild(svgZoomControls);
  return this.zoomControls_.init(bottom);
};

/**
 * Add a flyout.
 * @private
 */
Blockly.WorkspaceSvg.prototype.addFlyout_ = function() {
  var workspaceOptions = {
    disabledPatternId: this.options.disabledPatternId,
    parentWorkspace: this,
    RTL: this.RTL,
    horizontalLayout: this.horizontalLayout,
    toolboxPosition: this.options.toolboxPosition,
  };
  /** @type {Blockly.Flyout} */
  this.flyout_ = new Blockly.Flyout(workspaceOptions);
  this.flyout_.autoClose = false;
  var svgFlyout = this.flyout_.createDom();
  this.svgGroup_.insertBefore(svgFlyout, this.svgBlockCanvas_);
};

/**
 * Resize this workspace and its containing objects.
 */
Blockly.WorkspaceSvg.prototype.resize = function() {
  if (this.toolbox_) {
    this.toolbox_.position();
  }
  if (this.flyout_) {
    this.flyout_.position();
  }
  if (this.trashcan) {
    this.trashcan.position();
  }
  if (this.zoomControls_) {
    this.zoomControls_.position();
  }
  if (this.scrollbar) {
    this.scrollbar.resize();
  }
};

/**
 * Get the SVG element that forms the drawing surface.
 * @return {!Element} SVG element.
 */
Blockly.WorkspaceSvg.prototype.getCanvas = function() {
  return this.svgBlockCanvas_;
};

/**
 * Get the SVG element that forms the bubble surface.
 * @return {!SVGGElement} SVG element.
 */
Blockly.WorkspaceSvg.prototype.getBubbleCanvas = function() {
  return this.svgBubbleCanvas_;
};

/**
 * Get the SVG element that contains this workspace.
 * @return {!Element} SVG element.
 */
Blockly.WorkspaceSvg.prototype.getParentSvg = function() {
  if (this.cachedParentSvg_) {
    return this.cachedParentSvg_;
  }
  var element = this.svgGroup_;
  while (element) {
    if (element.tagName == 'svg') {
      this.cachedParentSvg_ = element;
      return element;
    }
    element = element.parentNode;
  }
  return null;
};

/**
 * Translate this workspace to new coordinates.
 * @param {number} x Horizontal translation.
 * @param {number} y Vertical translation.
 */
Blockly.WorkspaceSvg.prototype.translate = function(x, y) {
  var translation = 'translate(' + x + ',' + y + ') ' +
      'scale(' + this.scale + ')';
  this.svgBlockCanvas_.setAttribute('transform', translation);
  this.svgBubbleCanvas_.setAttribute('transform', translation);
};

/**
 * Returns the horizontal offset of the workspace.
 * Intended for LTR/RTL compatibility in XML.
 * @return {number} Width.
 */
Blockly.WorkspaceSvg.prototype.getWidth = function() {
  var metrics = this.getMetrics();
  return metrics ? metrics.viewWidth / this.scale : 0;
};

/**
 * Toggles the visibility of the workspace.
 * Currently only intended for main workspace.
 * @param {boolean} isVisible True if workspace should be visible.
 */
Blockly.WorkspaceSvg.prototype.setVisible = function(isVisible) {
  this.getParentSvg().style.display = isVisible ? 'block' : 'none';
  if (this.toolbox_) {
    // Currently does not support toolboxes in mutators.
    this.toolbox_.HtmlDiv.style.display = isVisible ? 'block' : 'none';
  }
  if (isVisible) {
    this.render();
    if (this.toolbox_) {
      this.toolbox_.position();
    }
  } else {
    Blockly.hideChaff(true);
  }
};

/**
 * Render all blocks in workspace.
 */
Blockly.WorkspaceSvg.prototype.render = function() {
  // Generate list of all blocks.
  var blocks = this.getAllBlocks();
  // Render each block.
  for (var i = blocks.length - 1; i >= 0; i--) {
    blocks[i].render(false);
  }
};

/**
 * Turn the visual trace functionality on or off.
 * @param {boolean} armed True if the trace should be on.
 */
Blockly.WorkspaceSvg.prototype.traceOn = function(armed) {
  this.traceOn_ = armed;
  if (this.traceWrapper_) {
    Blockly.unbindEvent_(this.traceWrapper_);
    this.traceWrapper_ = null;
  }
  if (armed) {
    this.traceWrapper_ = Blockly.bindEvent_(this.svgBlockCanvas_,
        'blocklySelectChange', this, function() {this.traceOn_ = false});
  }
};

/**
 * Highlight a block in the workspace.
 * @param {?string} id ID of block to find.
 */
Blockly.WorkspaceSvg.prototype.highlightBlock = function(id) {
  if (this.traceOn_ && Blockly.dragMode_ != 0) {
    // The blocklySelectChange event normally prevents this, but sometimes
    // there is a race condition on fast-executing apps.
    this.traceOn(false);
  }
  if (!this.traceOn_) {
    return;
  }
  var block = null;
  if (id) {
    block = Blockly.Block.getById(id);
    if (!block) {
      return;
    }
  }
  // Temporary turn off the listener for selection changes, so that we don't
  // trip the monitor for detecting user activity.
  this.traceOn(false);
  // Select the current block.
  if (block) {
    block.select();
  } else if (Blockly.selected) {
    Blockly.selected.unselect();
  }
  // Restore the monitor for user activity after the selection event has fired.
  var thisWorkspace = this;
  setTimeout(function() {thisWorkspace.traceOn(true);}, 1);
};

/**
 * @param {?string} id ID of block to find.
 * @param {boolean} isGlowing Whether to glow the block.
 */
Blockly.WorkspaceSvg.prototype.glowBlock = function(id, isGlowing) {
  var block = null;
  if (id) {
    block = Blockly.Block.getById(id);
    if (!block) {
      return;
    }
  }
  if (block) {
    block.setGlow(isGlowing);
  }
};

/**
 * Paste the provided block onto the workspace.
 * @param {!Element} xmlBlock XML block element.
 */
Blockly.WorkspaceSvg.prototype.paste = function(xmlBlock) {
  if (!this.rendered || xmlBlock.getElementsByTagName('block').length >=
      this.remainingCapacity()) {
    return;
  }
  Blockly.terminateDrag_();  // Dragging while pasting?  No.
  Blockly.Events.disable();
  var block = Blockly.Xml.domToBlock(this, xmlBlock);
  // Move the duplicate to original position.
  var blockX = parseInt(xmlBlock.getAttribute('x'), 10);
  var blockY = parseInt(xmlBlock.getAttribute('y'), 10);
  if (!isNaN(blockX) && !isNaN(blockY)) {
    if (this.RTL) {
      blockX = -blockX;
    }
    // Offset block until not clobbering another block and not in connection
    // distance with neighbouring blocks.
    do {
      var collide = false;
      var allBlocks = this.getAllBlocks();
      for (var i = 0, otherBlock; otherBlock = allBlocks[i]; i++) {
        var otherXY = otherBlock.getRelativeToSurfaceXY();
        if (Math.abs(blockX - otherXY.x) <= 1 &&
            Math.abs(blockY - otherXY.y) <= 1) {
          collide = true;
          break;
        }
      }
      if (!collide) {
        // Check for blocks in snap range to any of its connections.
        var connections = block.getConnections_(false);
        for (var i = 0, connection; connection = connections[i]; i++) {
          var neighbour =
              connection.closest(Blockly.SNAP_RADIUS, blockX, blockY);
          if (neighbour.connection) {
            collide = true;
            break;
          }
        }
      }
      if (collide) {
        if (this.RTL) {
          blockX -= Blockly.SNAP_RADIUS;
        } else {
          blockX += Blockly.SNAP_RADIUS;
        }
        blockY += Blockly.SNAP_RADIUS * 2;
      }
    } while (collide);
    block.moveBy(blockX, blockY);
  }
  Blockly.Events.enable();
  if (Blockly.Events.isEnabled() && !block.isShadow()) {
    Blockly.Events.fire(new Blockly.Events.Create(block));
  }
  block.select();
};

/**
 * Make a list of all the delete areas for this workspace.
 */
Blockly.WorkspaceSvg.prototype.recordDeleteAreas = function() {
  if (this.trashcan) {
    this.deleteAreaTrash_ = this.trashcan.getClientRect();
  } else {
    this.deleteAreaTrash_ = null;
  }
  if (this.flyout_) {
    this.deleteAreaToolbox_ = this.flyout_.getClientRect();
  } else if (this.toolbox_) {
    this.deleteAreaToolbox_ = this.toolbox_.getClientRect();
  } else {
    this.deleteAreaToolbox_ = null;
  }
};

/**
 * Is the mouse event over a delete area (toolbar or non-closing flyout)?
 * Opens or closes the trashcan and sets the cursor as a side effect.
 * @param {!Event} e Mouse move event.
 * @return {boolean} True if event is in a delete area.
 */
Blockly.WorkspaceSvg.prototype.isDeleteArea = function(e) {
  var isDelete = false;
  var xy = new goog.math.Coordinate(e.clientX, e.clientY);
  if (this.deleteAreaTrash_) {
    if (this.deleteAreaTrash_.contains(xy)) {
      this.trashcan.setOpen_(true);
      Blockly.Css.setCursor(Blockly.Css.Cursor.DELETE);
      return true;
    }
    this.trashcan.setOpen_(false);
  }
  if (this.deleteAreaToolbox_) {
    if (this.deleteAreaToolbox_.contains(xy)) {
      Blockly.Css.setCursor(Blockly.Css.Cursor.DELETE);
      return true;
    }
  }
  Blockly.Css.setCursor(Blockly.Css.Cursor.CLOSED);
  return false;
};

/**
 * Handle a mouse-down on SVG drawing surface.
 * @param {!Event} e Mouse down event.
 * @private
 */
Blockly.WorkspaceSvg.prototype.onMouseDown_ = function(e) {
  this.markFocused();
  if (Blockly.isTargetInput_(e)) {
    return;
  }
  Blockly.svgResize(this);
  Blockly.terminateDrag_();  // In case mouse-up event was lost.
  Blockly.hideChaff();
  var isTargetWorkspace = e.target && e.target.nodeName &&
      (e.target.nodeName.toLowerCase() == 'svg' ||
       e.target == this.svgBackground_);
  if (isTargetWorkspace && Blockly.selected && !this.options.readOnly) {
    // Clicking on the document clears the selection.
    Blockly.selected.unselect();
  }
  if (Blockly.isRightButton(e)) {
    // Right-click.
    this.showContextMenu_(e);
  } else if (this.scrollbar) {
    Blockly.removeAllRanges();
    // If the workspace is editable, only allow scrolling when gripping empty
    // space.  Otherwise, allow scrolling when gripping anywhere.
    this.isScrolling = true;
    // Record the current mouse position.
    this.startDragMouseX = e.clientX;
    this.startDragMouseY = e.clientY;
    this.startDragMetrics = this.getMetrics();
    this.startScrollX = this.scrollX;
    this.startScrollY = this.scrollY;

    // If this is a touch event then bind to the mouseup so workspace drag mode
    // is turned off and double move events are not performed on a block.
    // See comment in inject.js Blockly.init_ as to why mouseup events are
    // bound to the document instead of the SVG's surface.
    if ('mouseup' in Blockly.bindEvent_.TOUCH_MAP) {
      Blockly.onTouchUpWrapper_ = Blockly.onTouchUpWrapper_ || [];
      Blockly.onTouchUpWrapper_ = Blockly.onTouchUpWrapper_.concat(
          Blockly.bindEvent_(document, 'mouseup', null, Blockly.onMouseUp_));
    }
    Blockly.onMouseMoveWrapper_ = Blockly.onMouseMoveWrapper_ || [];
    Blockly.onMouseMoveWrapper_ = Blockly.onMouseMoveWrapper_.concat(
        Blockly.bindEvent_(document, 'mousemove', null, Blockly.onMouseMove_));
  }
  // This event has been handled.  No need to bubble up to the document.
  e.stopPropagation();
};

/**
 * Start tracking a drag of an object on this workspace.
 * @param {!Event} e Mouse down event.
 * @param {number} x Starting horizontal location of object.
 * @param {number} y Starting vertical location of object.
 */
Blockly.WorkspaceSvg.prototype.startDrag = function(e, x, y) {
  // Record the starting offset between the bubble's location and the mouse.
  var point = Blockly.mouseToSvg(e, this.getParentSvg());
  // Fix scale of mouse event.
  point.x /= this.scale;
  point.y /= this.scale;
  this.dragDeltaX_ = x - point.x;
  this.dragDeltaY_ = y - point.y;
};

/**
 * Track a drag of an object on this workspace.
 * @param {!Event} e Mouse move event.
 * @return {!goog.math.Coordinate} New location of object.
 */
Blockly.WorkspaceSvg.prototype.moveDrag = function(e) {
  var point = Blockly.mouseToSvg(e, this.getParentSvg());
  // Fix scale of mouse event.
  point.x /= this.scale;
  point.y /= this.scale;
  var x = this.dragDeltaX_ + point.x;
  var y = this.dragDeltaY_ + point.y;
  return new goog.math.Coordinate(x, y);
};

/**
 * Handle a mouse-wheel on SVG drawing surface.
 * @param {!Event} e Mouse wheel event.
 * @private
 */
Blockly.WorkspaceSvg.prototype.onMouseWheel_ = function(e) {
  // TODO: Remove terminateDrag and compensate for coordinate skew during zoom.
  Blockly.terminateDrag_();
  var delta = e.deltaY > 0 ? -1 : 1;
  var position = Blockly.mouseToSvg(e, this.getParentSvg());
  this.zoom(position.x, position.y, delta);
  e.preventDefault();
};

/**
 * Clean up the workspace by ordering all the blocks in a column.
 * @private
 */
Blockly.WorkspaceSvg.prototype.cleanUp_ = function() {
  var topBlocks = this.getTopBlocks(true);
  var cursorY = 0;
  for (var i = 0, block; block = topBlocks[i]; i++) {
    var xy = block.getRelativeToSurfaceXY();
    block.moveBy(-xy.x, cursorY - xy.y);
    block.snapToGrid();
    cursorY = block.getRelativeToSurfaceXY().y +
        block.getHeightWidth().height + Blockly.BlockSvg.MIN_BLOCK_Y;
  }
  // Fire an event to allow scrollbars to resize.
  Blockly.fireUiEvent(window, 'resize');
};

/**
 * Show the context menu for the workspace.
 * @param {!Event} e Mouse event.
 * @private
 */
Blockly.WorkspaceSvg.prototype.showContextMenu_ = function(e) {
  if (this.options.readOnly || this.isFlyout) {
    return;
  }
  var menuOptions = [];
  var topBlocks = this.getTopBlocks(true);
  // Option to clean up blocks.
  var cleanOption = {};
  cleanOption.text = Blockly.Msg.CLEAN_UP;
  cleanOption.enabled = topBlocks.length > 1;
  cleanOption.callback = this.cleanUp_.bind(this);
  menuOptions.push(cleanOption);

  // Add a little animation to collapsing and expanding.
  var DELAY = 10;
  if (this.options.collapse) {
    var hasCollapsedBlocks = false;
    var hasExpandedBlocks = false;
    for (var i = 0; i < topBlocks.length; i++) {
      var block = topBlocks[i];
      while (block) {
        if (block.isCollapsed()) {
          hasCollapsedBlocks = true;
        } else {
          hasExpandedBlocks = true;
        }
        block = block.getNextBlock();
      }
    }

    /*
     * Option to collapse or expand top blocks
     * @param {boolean} shouldCollapse Whether a block should collapse.
     * @private
     */
    var toggleOption = function(shouldCollapse) {
      var ms = 0;
      for (var i = 0; i < topBlocks.length; i++) {
        var block = topBlocks[i];
        while (block) {
          setTimeout(block.setCollapsed.bind(block, shouldCollapse), ms);
          block = block.getNextBlock();
          ms += DELAY;
        }
      }
    };

    // Option to collapse top blocks.
    var collapseOption = {enabled: hasExpandedBlocks};
    collapseOption.text = Blockly.Msg.COLLAPSE_ALL;
    collapseOption.callback = function() {
      toggleOption(true);
    };
    menuOptions.push(collapseOption);

    // Option to expand top blocks.
    var expandOption = {enabled: hasCollapsedBlocks};
    expandOption.text = Blockly.Msg.EXPAND_ALL;
    expandOption.callback = function() {
      toggleOption(false);
    };
    menuOptions.push(expandOption);
  }

  // Option to delete all blocks.
  // Count the number of blocks that are deletable.
  var deleteList = [];
  function addDeletableBlocks(block) {
    if (block.isDeletable()) {
      deleteList = deleteList.concat(block.getDescendants());
    } else {
      var children = block.getChildren();
      for (var i = 0; i < children.length; i++) {
        addDeletableBlocks(children[i]);
      }
    }
  }
  for (var i = 0; i < topBlocks.length; i++) {
    addDeletableBlocks(topBlocks[i]);
  }
  var deleteOption = {
    text: deleteList.length <= 1 ? Blockly.Msg.DELETE_BLOCK :
        Blockly.Msg.DELETE_X_BLOCKS.replace('%1', String(deleteList.length)),
    enabled: deleteList.length > 0,
    callback: function() {
      if (deleteList.length < 2 ||
          window.confirm(Blockly.Msg.DELETE_ALL_BLOCKS.replace('%1',
          String(deleteList.length)))) {
        deleteNext();
      }
    }
  };
  function deleteNext() {
    var block = deleteList.shift();
    if (block) {
      if (block.workspace) {
        block.dispose(false, true);
        setTimeout(deleteNext, DELAY);
      } else {
        deleteNext();
      }
    }
  }
  menuOptions.push(deleteOption);

  Blockly.ContextMenu.show(e, menuOptions, this.RTL);
};

/**
 * Set the callback for playing audio. The application should provide
 * a function that plays the called audio sound.
 * @param {!Function} func Function to call.
 */
Blockly.Workspace.prototype.setAudioCallback = function(func) {
  this.audioCallback_ = func;
};

/**
 * Play an audio file at specified value.
 * @param {string} name Name of sound.
 */
Blockly.WorkspaceSvg.prototype.playAudio = function(name) {
  if (this.audioCallback_) {
      this.audioCallback_(name);
  }
};

/**
 * Modify the block tree on the existing toolbox.
 * @param {Node|string} tree DOM tree of blocks, or text representation of same.
 */
Blockly.WorkspaceSvg.prototype.updateToolbox = function(tree) {
  tree = Blockly.parseToolboxTree_(tree);
  if (!tree) {
    if (this.options.languageTree) {
      throw 'Can\'t nullify an existing toolbox.';
    }
    return;  // No change (null to null).
  }
  if (!this.options.languageTree) {
    throw 'Existing toolbox is null.  Can\'t create new toolbox.';
  }
  if (tree.getElementsByTagName('category').length) {
    if (!this.toolbox_) {
      throw 'Existing toolbox has no categories.  Can\'t change mode.';
    }
    this.options.languageTree = tree;
    this.toolbox_.populate_(tree);
    this.toolbox_.addColour_();
  } else {
    if (!this.flyout_) {
      throw 'Existing toolbox has categories.  Can\'t change mode.';
    }
    this.options.languageTree = tree;
    this.flyout_.show(tree.childNodes);
  }
};

/**
 * Mark this workspace as the currently focused main workspace.
 */
Blockly.WorkspaceSvg.prototype.markFocused = function() {
  if (this.options.parentWorkspace) {
    this.options.parentWorkspace.markFocused();
  } else {
    Blockly.mainWorkspace = this;
  }
};

/**
 * Zooming the blocks centered in (x, y) coordinate with zooming in or out.
 * @param {number} x X coordinate of center.
 * @param {number} y Y coordinate of center.
 * @param {number} type Type of zooming (-1 zooming out and 1 zooming in).
 */
Blockly.WorkspaceSvg.prototype.zoom = function(x, y, type) {
  var speed = this.options.zoomOptions.scaleSpeed;
  var metrics = this.getMetrics();
  var center = this.getParentSvg().createSVGPoint();
  center.x = x;
  center.y = y;
  center = center.matrixTransform(this.getCanvas().getCTM().inverse());
  x = center.x;
  y = center.y;
  var canvas = this.getCanvas();
  // Scale factor.
  var scaleChange = (type == 1) ? speed : 1 / speed;
  // Clamp scale within valid range.
  var newScale = this.scale * scaleChange;
  if (newScale > this.options.zoomOptions.maxScale) {
    scaleChange = this.options.zoomOptions.maxScale / this.scale;
  } else if (newScale < this.options.zoomOptions.minScale) {
    scaleChange = this.options.zoomOptions.minScale / this.scale;
  }
  var matrix = canvas.getCTM()
      .translate(x * (1 - scaleChange), y * (1 - scaleChange))
      .scale(scaleChange);
  // newScale and matrix.a should be identical (within a rounding error).
  if (this.scale == matrix.a) {
    return;  // No change in zoom.
  }
  this.scale = matrix.a;
  this.scrollX = matrix.e - metrics.absoluteLeft;
  this.scrollY = matrix.f - metrics.absoluteTop;
  this.updateGridPattern_();
  if (this.scrollbar) {
    this.scrollbar.resize();
  } else {
    this.translate(0, 0);
  }
  Blockly.hideChaff(false);
  if (this.flyout_) {
    // No toolbox, resize flyout.
    this.flyout_.reflow();
  }
};

/**
 * Zooming the blocks centered in the center of view with zooming in or out.
 * @param {number} type Type of zooming (-1 zooming out and 1 zooming in).
 */
Blockly.WorkspaceSvg.prototype.zoomCenter = function(type) {
  var metrics = this.getMetrics();
  var x = metrics.viewWidth / 2;
  var y = metrics.viewHeight / 2;
  this.zoom(x, y, type);
};

/**
 * Zoom the blocks to fit in the workspace if possible.
 */
Blockly.WorkspaceSvg.prototype.zoomToFit = function() {
  var workspaceBBox = this.svgBackground_.getBBox();
  var blocksBBox = this.svgBlockCanvas_.getBBox();
  var workspaceWidth = workspaceBBox.width - this.toolbox_.width -
      Blockly.Scrollbar.scrollbarThickness;
  var workspaceHeight = workspaceBBox.height -
      Blockly.Scrollbar.scrollbarThickness;
  var blocksWidth = blocksBBox.width;
  var blocksHeight = blocksBBox.height;
  if (blocksWidth == 0) {
    return;  // Prevents zooming to infinity.
  }
  var ratioX = workspaceWidth / blocksWidth;
  var ratioY = workspaceHeight / blocksHeight;
  var ratio = Math.min(ratioX, ratioY);
  var speed = this.options.zoomOptions.scaleSpeed;
  var numZooms = Math.floor(Math.log(ratio) / Math.log(speed));
  var newScale = Math.pow(speed, numZooms);
  if (newScale > this.options.zoomOptions.maxScale) {
    newScale = this.options.zoomOptions.maxScale;
  } else if (newScale < this.options.zoomOptions.minScale) {
    newScale = this.options.zoomOptions.minScale;
  }
  this.scale = newScale;
  this.updateGridPattern_();
  this.scrollbar.resize();
  Blockly.hideChaff(false);
  if (this.flyout_) {
    // No toolbox, resize flyout.
    this.flyout_.reflow();
  }
  // Center the workspace.
  var metrics = this.getMetrics();
  this.scrollbar.set((metrics.contentWidth - metrics.viewWidth) / 2,
      (metrics.contentHeight - metrics.viewHeight) / 2);
};

/**
 * Reset zooming and dragging.
 * @param {!Event} e Mouse down event.
 */
Blockly.WorkspaceSvg.prototype.zoomReset = function(e) {
  this.scale = 1;
  this.updateGridPattern_();
  Blockly.hideChaff(false);
  if (this.flyout_) {
    // No toolbox, resize flyout.
    this.flyout_.reflow();
  }
  // Zoom level has changed, update the scrollbars.
  if (this.scrollbar) {
    this.scrollbar.resize();
  }
  // Center the workspace.
  var metrics = this.getMetrics();
  if (this.scrollbar) {
    this.scrollbar.set((metrics.contentWidth - metrics.viewWidth) / 2,
        (metrics.contentHeight - metrics.viewHeight) / 2);
  } else {
    this.translate(0, 0);
  }
  // This event has been handled.  Don't start a workspace drag.
  e.stopPropagation();
};

/**
 * Updates the grid pattern.
 * @private
 */
Blockly.WorkspaceSvg.prototype.updateGridPattern_ = function() {
  if (!this.options.gridPattern) {
    return;  // No grid.
  }
  // MSIE freaks if it sees a 0x0 pattern, so set empty patterns to 100x100.
  var safeSpacing = (this.options.gridOptions['spacing'] * this.scale) || 100;
  this.options.gridPattern.setAttribute('width', safeSpacing);
  this.options.gridPattern.setAttribute('height', safeSpacing);
  var half = Math.floor(this.options.gridOptions['spacing'] / 2) + 0.5;
  var start = half - this.options.gridOptions['length'] / 2;
  var end = half + this.options.gridOptions['length'] / 2;
  var line1 = this.options.gridPattern.firstChild;
  var line2 = line1 && line1.nextSibling;
  half *= this.scale;
  start *= this.scale;
  end *= this.scale;
  if (line1) {
    line1.setAttribute('stroke-width', this.scale);
    line1.setAttribute('x1', start);
    line1.setAttribute('y1', half);
    line1.setAttribute('x2', end);
    line1.setAttribute('y2', half);
  }
  if (line2) {
    line2.setAttribute('stroke-width', this.scale);
    line2.setAttribute('x1', half);
    line2.setAttribute('y1', start);
    line2.setAttribute('x2', half);
    line2.setAttribute('y2', end);
  }
};

// Export symbols that would otherwise be renamed by Closure compiler.
Blockly.WorkspaceSvg.prototype['setVisible'] =
    Blockly.WorkspaceSvg.prototype.setVisible;
