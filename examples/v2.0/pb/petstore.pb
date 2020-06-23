
2.0 
Swagger Petstore1.0.02
MITpetstore.swagger.io"/v1*http2application/json:application/jsonBº¯
/pets¥ª
petsList all pets*listPetsBT
RPNquery.How many items to return at one time (max 100)"limit2integer:int32J²
v
200o
m
An paged array of pets

#/definitions/Pets;
9
x-next/
string’$A link to the next page of responses
8
default-
+
unexpected error

#/definitions/Error"v
petsCreate a pet*
createPetsJT

201

Null response
8
default-
+
unexpected error

#/definitions/Error…
/pets/{petId}óð
petsInfo for a specific pet*showPetByIdB<
:8"6pathThe id of the pet to retrieve"petId*stringJƒ
G
200@
>
$Expected response to a valid request

#/definitions/Pets
8
default-
+
unexpected error

#/definitions/ErrorJÜ
Z
PetSšidšnameÊD

idint64²	
integer

name²
string

tag²
string
*
Pets"²
arrayº

#/definitions/Pet
R
ErrorIšcodešmessageÊ5

codeint32²	
integer

message²
string