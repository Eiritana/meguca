/*
 * OP and thread related logic
 */

var $ = require('jquery'),
	_ = require('underscore'),
	Backbone = require('backbone'),
	imager = require('./imager'),
	main = require('../main'),
	postCommon = require('./common'),
	state = require('../state');

var Section = module.exports = Backbone.View.extend({
	tagName: 'section',

	initialize: function () {
		// On the live page only
		if (this.$el.is(':empty') && state.page.get('live'))
			this.render();
		this.listenTo(this.model, {
			'change:locked': this.renderLocked,
			destroy: this.remove
		});
		this.listenToOnce(this.model, {
			'add': this.renderRelativeTime
		});
		this.initCommon();
	},

	render: function() {
		main.oneeSama.links = this.model.get('links');
		this.setElement(main.oneeSama.monomono(this.model.attributes).join(''));
		this.$el.insertAfter(main.$threads.children('aside:first'));
		// Insert reply box into the new thread
		var $reply = $(main.oneeSama.replyBox());
		if (state.ownPosts.hasOwnProperty(this.model.get('num')) || main.postForm)
			$reply.hide();
		this.$el.after($reply, '<hr>');
		return this;
	},

	renderHide: function (model, hide) {
		this.$el.next('hr').andSelf().toggle(!hide);
	},

	renderLocked: function (model, locked) {
		this.$el.toggleClass('locked', !!locked);
	},

	remove: function () {
		this.$el.next('hr').andSelf().remove();
		this.stopListening();
		return this;
	},

	/*
	 Remove the top reply on board pages, if over limit, when a new reply is
	 added
	 */
	shiftReplies: function() {

	}
});

// Extend with common mixins
_.extend(Section.prototype, imager.Hidamari, postCommon);
