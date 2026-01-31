from django.contrib import admin
from django import forms
from adminsortable2.admin import SortableAdminBase, SortableInlineAdminMixin
from .models import Events, EventBlock, MarysRequests, ExternalSources


class EventBlockInlineForm(forms.ModelForm):
    class Meta:
        model = EventBlock
        fields = '__all__'
        widgets = {
            'title': forms.TextInput(attrs={'style': 'width: 100%; min-width: 150px;'}),
        }


class EventBlockInline(SortableInlineAdminMixin, admin.TabularInline):
    model = EventBlock
    form = EventBlockInlineForm
    extra = 1
    fields = ['language', 'title', 'content', 'ordering']
    ordering = ['ordering']


@admin.register(Events)
class EventsAdmin(SortableAdminBase, admin.ModelAdmin):
    list_display = ['name', 'category', 'years', 'block_count']
    search_fields = ['name', 'description']
    list_filter = ['category', 'years']
    inlines = [EventBlockInline]

    @admin.display(description='Blocks')
    def block_count(self, obj):
        return obj.blocks.count()


@admin.register(EventBlock)
class EventBlockAdmin(admin.ModelAdmin):
    list_display = ['event', 'language', 'title', 'ordering', 'updated_at']
    list_filter = ['language', 'event']
    search_fields = ['title', 'content']
    ordering = ['event', 'ordering']
