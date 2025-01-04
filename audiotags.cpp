/***************************************************************************
    copyright            : (C) 2003 by Scott Wheeler
    email                : wheeler@kde.org
 ***************************************************************************/

/***************************************************************************
    copyright            : (C) 2014 by Nick Sellen
    email                : code@nicksellen.co.uk
 ***************************************************************************/

/***************************************************************************
 *   This library is free software; you can redistribute it and/or modify  *
 *   it  under the terms of the GNU Lesser General Public License version  *
 *   2.1 as published by the Free Software Foundation.                     *
 *                                                                         *
 *   This library is distributed in the hope that it will be useful, but   *
 *   WITHOUT ANY WARRANTY; without even the implied warranty of            *
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU     *
 *   Lesser General Public License for more details.                       *
 *                                                                         *
 *   You should have received a copy of the GNU Lesser General Public      *
 *   License along with this library; if not, write to the Free Software   *
 *   Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA  02111-1307  *
 *   USA                                                                   *
 ***************************************************************************/

#include "audiotags.h"

static bool unicodeStrings = true;

TagLib_File *audiotags_file_new(const char *filename)
{
  TagLib_File *fr = taglib_file_new(filename);
  if (fr == NULL || !taglib_file_is_valid(fr)) {
    taglib_file_free(fr);
  }

  return fr;
}

TagLib_File *audiotags_file_memory(const char *data, unsigned int length) {
  TagLib_IOStream *ioStream = taglib_memory_iostream_new(data, length);
  TagLib_File *fr = taglib_file_new_iostream(ioStream);
  if (fr == NULL || !taglib_file_is_valid(fr)) {
    taglib_file_free(fr);
    taglib_iostream_free(ioStream);
    return NULL;
  }

  return fr;
}

void audiotags_file_properties(const TagLib_File *fileRef, int id)
{
  // From https://github.com/taglib/taglib/blob/master/tests/test_tag_c.cpp
  if (char **keys = taglib_property_keys(fileRef)) {
    char **keyPtr = keys;
    
    while (*keyPtr) {
      char **values = taglib_property_get(fileRef, *keyPtr);
      char **valuePtr = values;
      while (*valuePtr) {
        goTagPut(id, *keyPtr, *valuePtr);
        *valuePtr++;
      }

      taglib_property_free(values);
      *keyPtr++;
    }

    taglib_property_free(keys);
  }
}

bool audiotags_clear_properties(TagLib_File *fileRef)
{
  if (char **keys = taglib_property_keys(fileRef)) {
    char **keyPtr = keys;
    
    while (*keyPtr) {
      taglib_property_set(fileRef, *keyPtr, NULL);
      *keyPtr++;
    }

    taglib_property_free(keys);
  }

  return true;
}

bool audiotags_write_properties(TagLib_File *fileRef, unsigned int len, const char *fields_c[], const char *values_c[])
{
  for (unsigned int i = 0; i < len; i++) {
    taglib_property_set(fileRef, fields_c[i], values_c[i]);
  }

  return taglib_file_save(fileRef);
}

const TagLib_AudioProperties *audiotags_file_audioproperties(const TagLib_File *fileRef)
{
  return taglib_file_audioproperties(fileRef);
}

int audiotags_audioproperties_length(const TagLib_AudioProperties *audioProperties)
{
  return taglib_audioproperties_length(audioProperties);
}

int audiotags_audioproperties_bitrate(const TagLib_AudioProperties *audioProperties)
{
  return taglib_audioproperties_bitrate(audioProperties);
}

int audiotags_audioproperties_samplerate(const TagLib_AudioProperties *audioProperties)
{
  return taglib_audioproperties_samplerate(audioProperties);
}

int audiotags_audioproperties_channels(const TagLib_AudioProperties *audioProperties)
{
  return taglib_audioproperties_channels(audioProperties);
}

bool audiotags_read_picture(TagLib_File *fileRef, int id)
{
  TagLib_Complex_Property_Attribute*** properties = taglib_complex_property_get(fileRef, "PICTURE");
  if (!properties)
    return false;

  // TODO: Seems to be no way to verify picture
  TagLib_Complex_Property_Picture_Data picture;
  taglib_picture_from_complex_property(properties, &picture);

  goPutImage(id, picture.data, picture.size);

  return true;
}

bool audiotags_write_picture(TagLib_File *fileRef, const char *data, unsigned int length, int w, int h, const char *mime)
{
  TAGLIB_COMPLEX_PROPERTY_PICTURE(prop, data, length, "Written by go-taglib", mime, "Front Cover");

  if (!taglib_complex_property_set(fileRef, "PICTURE", prop))
    return false;

  return taglib_file_save(fileRef);
}

bool audiotags_remove_pictures(TagLib_File *fileRef)
{
  // https://github.com/taglib/taglib/issues/938#issuecomment-1773688643
  if (!taglib_complex_property_set(fileRef, "PICTURE", NULL))
    return false;

  return taglib_file_save(fileRef); 
}
