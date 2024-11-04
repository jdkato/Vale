1. Enter the project details.
   - In the **Project name** field, enter `animals`.

## Create a Django application

1. Create a model to define an Animal. In `animals/models.py`:

   ```python
   from django.db import models
   class Animal(models.Model):
       name = models.CharField(max_length=200)
       number_of_legs = models.IntegerField(default=2)
       dangerous = models.BooleanField(default=False)
   ```