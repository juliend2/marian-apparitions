# This is an auto-generated Django model module.
# You'll have to do the following manually to clean this up:
#   * Make sure each ForeignKey and OneToOneField has `on_delete` set to the desired behavior
#   * Remove `managed = False` lines if you wish to allow Django to create, modify, and delete the table
# Feel free to rename the models, but don't rename db_table values or field names.
from django.db import models


class Events(models.Model):
    id = models.AutoField(primary_key=True)
    category = models.TextField(blank=True, null=True)
    name = models.TextField(blank=True, null=True)
    description = models.TextField(blank=True, null=True)
    wikipedia_section_title = models.TextField(blank=True, null=True)
    image_filename = models.TextField(blank=True, null=True)
    years = models.TextField(blank=True, null=True)
    slug = models.TextField(blank=True, null=True)
    country = models.TextField(blank=True, null=True)

    class Meta:
        managed = False
        db_table = 'events'

    def __str__(self):
        return self.name or f"Event {self.id}"

class MarysRequests(models.Model):
    id = models.AutoField(primary_key=True)
    event = models.ForeignKey(Events, on_delete=models.CASCADE, blank=True, null=True, db_column='event_id')
    request = models.TextField(blank=True, null=True)

    class Meta:
        managed = False
        db_table = 'marys_requests'

class ExternalSources(models.Model):
    id = models.AutoField(primary_key=True)
    event = models.ForeignKey(Events, on_delete=models.CASCADE, db_column='event_id')
    source_url = models.TextField(blank=True, null=True)

    class Meta:
        managed = False
        db_table = 'external_sources'

class EventBlock(models.Model):
    id = models.AutoField(primary_key=True)
    event = models.ForeignKey(Events, on_delete=models.CASCADE, related_name='blocks', db_column='event_id')
    language = models.CharField(max_length=10, default='en')
    title = models.TextField(blank=True, null=True)
    content = models.TextField(blank=True, null=True)
    ordering = models.IntegerField(default=0)
    church_authority = models.CharField(max_length=100, blank=True, null=True)
    authority_position = models.CharField(max_length=50, blank=True, null=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        managed = True
        db_table = 'event_blocks'
        ordering = ['ordering', 'id']

    def __str__(self):
        return f"{self.event.name} - Block {self.ordering}: {self.title or '(untitled)'}"
