from __future__ import absolute_import, unicode_literals

from django.core.exceptions import ValidationError
from django.db import models


def get_sentinel_user():
    return get_user_model().objects.get_or_create(
        username="pontoon-deleted-user", email="pontoon-deleted-user@mozilla.com"
    )[0]


class ActionLog(models.Model):
    ACTIONS_TYPES = (
        # A translation has been created.
        ("translation:created", "Translation created"),
        # A translation has been deleted.
        ("translation:deleted", "Translation deleted"),
        # A translation has been approved.
        ("translation:approved", "Translation approved"),
        # A translation has been unapproved.
        ("translation:unapproved", "Translation unapproved"),
        # A translation has been rejected.
        ("translation:rejected", "Translation rejected"),
        # A translation has been unrejected.
        ("translation:unrejected", "Translation unrejected"),
        # A comment has been added.
        ("comment:added", "Comment added"),
    )

    action_type = models.CharField(max_length=50, choices=ACTIONS_TYPES)
    created_at = models.DateTimeField(auto_now_add=True)
    performed_by = models.ForeignKey(
        "auth.User", models.SET(get_sentinel_user), related_name="actions"
    )

    # Used to track on what translation related actions apply.
    translation = models.ForeignKey(
        "base.Translation", models.CASCADE, blank=True, null=True,
    )

    # Used when a translation has been deleted.
    entity = models.ForeignKey("base.Entity", models.CASCADE, blank=True, null=True,)
    locale = models.ForeignKey("base.Locale", models.CASCADE, blank=True, null=True,)

    def validate_action_type_choice(self):
        valid_types = [t[0] for t in self.ACTIONS_TYPES]
        if self.action_type not in valid_types:
            raise ValidationError(
                'Action type "{}" is not one of the permitted values: {}'.format(
                    self.action_type, ", ".join(valid_types)
                )
            )

    def validate_foreign_keys_per_action(self):
        if self.action_type == "translation:deleted" and (
            self.translation or not self.entity or not self.locale
        ):
            raise ValidationError(
                'For action "translation:deleted", `entity` and `locale` are required'
            )

        if self.action_type == "comment:added" and not (
            (self.translation and not self.locale and not self.entity)
            or (not self.translation and self.locale and self.entity)
        ):
            raise ValidationError(
                'For action type "comment:added", either `translation` or `locale` and `entity` are required'
            )

        if (
            self.action_type != "translation:deleted"
            and self.action_type != "comment:added"
        ) and (not self.translation or self.entity or self.locale):
            raise ValidationError(
                'Only `translation` is accepted for action type "{}"'.format(
                    self.action_type
                )
            )

    def save(self, *args, **kwargs):
        self.validate_action_type_choice()
        self.validate_foreign_keys_per_action()

        super(ActionLog, self).save(*args, **kwargs)
